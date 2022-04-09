package ar8t

// TODO: testing

type gf8 struct {
	v byte
}

func (f gf8) AddOrSub(other gf8) gf8 {
	return gf8{f.v ^ other.v}
}

func (f gf8) Mul(other gf8) gf8 {
	if f.v == 0 || other.v == 0 {
		return exp8[255]
	}

	logF, logOther := log8[f.v], log8[other.v]

	return exp8[(uint16(logF)+uint16(logOther))%255]

}

func (f gf8) Div(other gf8) gf8 {
	logF, logOther := log8[f.v], log8[other.v]
	diff := int16(logF) - int16(logOther)

	if diff < 0 {
		diff += 255
	}

	return exp8[diff%255]

}

type gf4 struct {
	v uint8
}

func (f gf4) AddOrSub(other gf4) gf4 {
	return gf4{f.v ^ other.v}
}

func (f gf4) Mul(other gf4) gf4 {
	if f.v == 0 || other.v == 0 {
		return exp4[15]
	}

	logF, logOther := log4[f.v], log4[other.v]

	return exp4[(uint16(logF)+uint16(logOther))%15]

}

func (f gf4) Div(other gf4) gf4 {
	logF, logOther := log4[f.v], log4[other.v]
	diff := int16(logF) - int16(logOther)

	if diff < 0 {
		diff += 15
	}

	return exp4[diff]

}
