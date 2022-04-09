package ar8t

import (
	"bytes"
)

// I don't like the file's name, we'll see about it

type QRLocation struct {
	TopLeft, TopRight, BottomLeft Point
	ModuleSize                    float64 //in pixels
	Version                       uint32  //1 .. 40
}

type QRData struct {
	/// QR Pixel Data in side x side pixels, stored in row major order. Using the provided index will convert into 1's and 0's.
	Data []byte

	/// Version of the QR Code, 1 being the smallest, 40 the largest
	Version uint32

	/// Side in pixels of the QR square
	Side uint32
}

func (d QRData) Index(x, y uint32) byte {
	index := y*d.Side + x
	if int(index) >= len(d.Data) {
		return 1
	}

	switch d.Data[index] {
	case 0:
		return 1
	default:
		return 0
	}
}

type ECLevel int

const (
	ECLevelLow ECLevel = iota
	ECLevelMedium
	ECLevelQuartile
	ECLevelHigh
)

type Chomp struct {
	bytes            *bytes.Reader
	bitsLeft         uint32
	currentByte      byte
	currentByteValid bool
	bitsLeftInByte   uint8
}

// NewChomp creates a chomper dependant on bytes
func NewChomp(b []byte) *Chomp {
	ch := &Chomp{}

	ch.bitsLeft = uint32(len(b) * 8)
	ch.bytes = bytes.NewReader(b)
	var err error
	ch.currentByte, err = ch.bytes.ReadByte()
	// otherwise it's EOF
	if err == nil {
		ch.bitsLeftInByte = 8
		ch.currentByteValid = true
	}

	return ch
}

func (ch *Chomp) Chomp(nBits uint8) (result byte, ok bool) {
	bitCount := nBits
	if bitCount > 8 && bitCount < 1 ||
		uint32(bitCount) > ch.bitsLeft {
		return
	}

	switch {
	case bitCount < ch.bitsLeftInByte:
		return ch.nibble(nBits)
	case bitCount == ch.bitsLeftInByte:
		if ch.currentByteValid {
			result = ch.currentByte >> (8 - ch.bitsLeftInByte)
		}

		ch.bitsLeft -= uint32(ch.bitsLeftInByte)
		ch.bitsLeftInByte = 0
		ch.currentByteValid = false

		var err error
		ch.currentByte, err = ch.bytes.ReadByte()
		if err == nil {
			ch.bitsLeftInByte = 8
			ch.currentByteValid = true
		}
		ok = true
		return
	default:
		bitsToGo := bitCount - ch.bitsLeftInByte
		if ch.currentByteValid {
			result = ch.currentByte >> (8 - ch.bitsLeftInByte) << bitsToGo
		}

		ch.bitsLeft -= uint32(ch.bitsLeftInByte)

		var err error
		ch.currentByte, err = ch.bytes.ReadByte()
		if err != nil {
			ch.currentByteValid = false
			ch.bitsLeftInByte = 0
			return
		}
		ch.bitsLeftInByte = 8

		nibble, _ := ch.nibble(bitsToGo)
		return result + nibble, true
	}

}

func (ch *Chomp) ChompUint16(nBits uint8) (uint16, bool) {
	var (
		result        uint16
		initialResult byte
		ok            bool
	)

	for nBits > 8 {
		initialResult, ok = ch.Chomp(8)
		if !ok {
			return 0, false
		}
		result = uint16(initialResult) << (nBits - 8)
		nBits -= 8
	}

	initialResult, ok = ch.Chomp(nBits)

	if !ok {
		return 0, false
	}

	return result + uint16(initialResult), true
}

func (ch *Chomp) nibble(nBits uint8) (result byte, ok bool) {
	if !ch.currentByteValid {
		return
	}

	result = ch.currentByte >> (8 - nBits)
	ch.currentByte <<= nBits

	ch.bitsLeftInByte -= nBits

	ok = true
	return
}
