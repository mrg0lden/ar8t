package ar8t

import "errors"

var maskArr = [15]byte{1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0}

func Format(data QRData) (ECLevel, QRMask, error) {
	format, err := format1(data)
	if err != nil {
		format, err = format2(data)
	}

	if err != nil {
		return 0, nil, err
	}

	correction, ok := errorCorrection(2*format[0] + format[1])
	if !ok {
		return 0, nil, errors.New("failed to correct errors")
	}

	mask, ok := mask(4*format[2] + 2*format[3] + format[4])
	if !ok {
		return 0, nil, errors.New("failed to obtain a mask")
	}

	return correction, mask, nil

}

func format1(data QRData) ([]byte, error) {
	format := []byte{}

	for x := 0; x < 9; x++ {
		if x == 6 {
			continue
		}

		format = append(format, data.Index(uint32(x), 8))
	}

	yFunc := revIndex(8)
	for y := 0; y < 8; y++ {
		y := yFunc(y)
		if y == 6 {
			continue
		}

		format = append(format, data.Index(8, uint32(y)))
	}

	for i := range format {
		format[i] ^= maskArr[i]
	}

	return correct(format)
}

func format2(data QRData) ([]byte, error) {
	format := []byte{}
	yFunc := revIndex(data.Side)
	for y := data.Side - 7; y < data.Side; y++ {
		y := yFunc(y)
		format = append(format, data.Index(8, y))
	}

	for x := data.Side - 8; x < data.Side; x++ {
		format = append(format, data.Index(uint32(x), 8))
	}

	for i := range format {
		format[i] ^= maskArr[i]
	}

	return correct(format)
}

// this function is a copy pasta,
// maybe can be written in a better way
func correct(format []byte) ([]byte, error) {
	s1 := gf4{}

	for i := range format {
		valIndex := len(format) - i - 1
		s1 = s1.AddOrSub(gf4{format[valIndex]}.Mul(exp4[i%15]))
	}

	if s1 == (gf4{}) {
		// syndrome == 0, no error detected
		return format, nil
	}

	var (
		s2 = s1.Mul(s1)
		s4 = s2.Mul(s2)

		s3, s5 gf4
	)

	for i := range format {
		index := len(format) - i - 1
		s3 = s3.AddOrSub(gf4{format[index]}.Mul(exp4[(3*i)%15]))
		s5 = s5.AddOrSub(gf4{format[index]}.Mul(exp4[(5*i)%15]))
	}

	sigma1 := s1
	sigma2 := s5.AddOrSub(s4.Mul(sigma1)).AddOrSub(s2.Mul(s3.AddOrSub(s2.Mul(sigma1))).Div(s3.AddOrSub(s1.Mul(s2))))
	sigma3 := s3.AddOrSub(s2.Mul(sigma1)).AddOrSub(s1.Mul(sigma2))

	errorPos := []byte{}

	for i := byte(0); i < 16; i++ {
		x := gf4{i}
		if sigma3.AddOrSub(sigma2.Mul(x)).
			AddOrSub(sigma1.Mul(x).Mul(x)).
			AddOrSub(x.Mul(x).Mul(x)) == (gf4{}) {
			log := log4[i]
			if log != 0 {
				errorPos = append(errorPos, log)
			}
		}
	}

	for _, err := range errorPos {
		if i := len(format) - int(err) - 1; i >= 0 {
			format[i] ^= 1
		}
	}

	s1 = gf4{}

	for i := range format {
		s1 = s1.AddOrSub(gf4{format[len(format)-i-1]}.Mul(exp4[i%15]))
	}

	if s1 == (gf4{}) {
		// syndrome == 0, no error detected
		return format, nil
	}

	return nil, errors.New("format information corrupted")

}

func errorCorrection(bits byte) (level ECLevel, ok bool) {
	ok = true
	switch bits {
	case 0b01:
		level = ECLevelLow
	case 0b00:
		level = ECLevelMedium
	case 0b11:
		level = ECLevelQuartile
	case 0b10:
		level = ECLevelHigh
	default:
		level, ok = 0, false
	}
	return
}

func mask(bits byte) (QRMask, bool) {
	switch bits {
	case 0b000:
		return qrMask(func(j, i uint32) bool { return (i+j)%2 == 0 })
	case 0b001:
		return qrMask(func(_, i uint32) bool { return i%2 == 0 })
	case 0b010:
		return qrMask(func(j, _ uint32) bool { return j%3 == 0 })
	case 0b011:
		return qrMask(func(j, i uint32) bool { return (i+j)%3 == 0 })
	case 0b100:
		return qrMask(func(j, i uint32) bool { return (i/2+j/3)%2 == 0 })
	case 0b101:
		return qrMask(func(j, i uint32) bool { return (i*j)%2+(i*j)%3 == 0 })
	case 0b110:
		return qrMask(func(j, i uint32) bool { return ((i*j)%2+(i*j)%3)%2 == 0 })
	case 0b111:
		return qrMask(func(j, i uint32) bool { return ((i*j)%3+(i+j)%2)%2 == 0 })
	default:
		return nil, false
	}
}

type formatMask = func(a, b uint32) bool

func qrMask(mask formatMask) (QRMask, bool) {
	return func(q QRData, i, j uint32) byte {
		maskVal := 0
		if mask(i, j) {
			maskVal = 1
		}
		return q.Index(i, j) ^ byte(maskVal)
	}, true
}
