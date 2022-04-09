package ar8t

import (
	"bytes"
	"fmt"
)

func Data(input []byte, version uint32) ([]byte, error) {
	chomp := NewChomp(input)
	result := bytes.Buffer{}

modeLoop:
	for mode, ok := chomp.Chomp(4); ok; mode, ok = chomp.Chomp(4) {
		var (
			data []byte
			err  error
		)

		switch mode {
		case 0b0001:
			data, err = numeric(chomp, version)
		case 0b0010:
			data, err = alphanumeric(chomp, version)
		case 0b0100:
			data, err = eightBit(chomp, version)
		case 0b0000:
			break modeLoop
		default:
			return nil, fmt.Errorf("mode %.4b not yet implemented", mode)
		}

		if err != nil {
			return nil, err
		}

		_, err = result.Write(data)
		if err != nil {
			return nil, err
		}
	}

	return result.Bytes(), nil
}

func numeric(chomp *Chomp, version uint32) ([]byte, error) {
	var lengthBits uint8
	switch {
	case version >= 1 && version <= 9:
		lengthBits = 10
	case version >= 10 && version <= 26:
		lengthBits = 12
	case version >= 27 && version <= 40:
		lengthBits = 14
	default:
		return nil, fmt.Errorf("unknown version %d", version)
	}

	length, ok := chomp.ChompUint16(lengthBits)
	if !ok {
		return nil, fmt.Errorf("could not read %d bits for numeric length", lengthBits)
	}

	result := bytes.Buffer{}

	for length > 0 {

		var (
			digits     uint16
			err        error
			fmtPattern string
			toBreak    bool
		)

		switch {
		case length >= 3:
			digits, err = readBitsUint16(chomp, 10)
			fmtPattern = "%.3d"
			length -= 3
		case length == 2:
			digits, err = readBitsUint16(chomp, 7)
			fmtPattern = "%.2d"
			toBreak = true
		case length == 1:
			digits, err = readBitsUint16(chomp, 4)
			fmtPattern = "%.1d"
			toBreak = true
		}

		if err != nil {
			return nil, err
		}

		result.WriteString(fmt.Sprintf(fmtPattern, digits))

		if toBreak {
			break
		}
	}

	return result.Bytes(), nil
}

var alphanumericArr = [45]byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
	'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', ' ', '$',
	'%', '*', '+', '-', '.', '/', ':',
}

func alphanumeric(chomp *Chomp, version uint32) ([]byte, error) {
	var lengthBits uint8
	switch {
	case version >= 1 && version <= 9:
		lengthBits = 9
	case version >= 10 && version <= 26:
		lengthBits = 11
	case version >= 27 && version <= 40:
		lengthBits = 13
	default:
		return nil, fmt.Errorf("unknown version %d", version)
	}

	length, ok := chomp.ChompUint16(lengthBits)
	if !ok {
		return nil, fmt.Errorf("could not read %d bits for alphanumeric length", lengthBits)
	}

	result := []byte{}

	for length > 0 {

		switch {
		case length >= 2:
			chars, err := readBitsUint16(chomp, 11)
			if err != nil {
				return nil, err
			}

			result = append(result, alphanumericArr[chars/45], alphanumericArr[chars%45])

			length -= 2
		case length == 1:
			chars, err := readBitsUint16(chomp, 6)
			if err != nil {
				return nil, err
			}

			result = append(result, alphanumericArr[chars])
			length -= 1 //essentially breaking the loop
		}
	}

	return result, nil
}

func eightBit(chomp *Chomp, version uint32) ([]byte, error) {
	var lengthBits uint8

	switch {
	case version >= 1 && version <= 9:
		lengthBits = 8
	case version >= 10 && version <= 26:
		fallthrough
	case version >= 27 && version <= 40:
		lengthBits = 16
	default:
		return nil, fmt.Errorf("unknown version %d", version)
	}

	length, ok := chomp.ChompUint16(lengthBits)
	if !ok {
		return nil, fmt.Errorf("could not read %d bits for 8bits length", length)
	}

	result := []byte{}

	for i := uint16(0); i < length; i++ {
		bits, err := readBits(chomp, 8)
		if err != nil {
			return nil, err
		}

		result = append(result, bits)
	}

	// we will not convert this input to a string
	// we will leave this choice for the end user of this
	//package

	return result, nil
}

func readBits(chomp *Chomp, nBits byte) (byte, error) {
	bits, ok := chomp.Chomp(nBits)
	if !ok {
		return 0, fmt.Errorf("could not read %d bits", nBits)
	}

	return bits, nil
}

func readBitsUint16(chomp *Chomp, nBits byte) (uint16, error) {
	bits, ok := chomp.ChompUint16(nBits)
	if !ok {
		return 0, fmt.Errorf("could not read %d bits", nBits)
	}

	return bits, nil
}
