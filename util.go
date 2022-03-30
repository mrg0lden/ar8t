package ar8t

import "image"

// I don't like the file's name, we'll see about it

type QRLocation struct {
	TopLeft, TopRight, BottomLeft image.Point
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

type ECLevel int

const (
	ECLevelLow ECLevel = iota
	ECLevelMedium
	ECLevelQuartile
	ECLevelHigh
)
