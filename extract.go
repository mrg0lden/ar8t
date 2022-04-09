package ar8t

import (
	"errors"
	"image"
	"math"
)

var errUnableToFindPattern = errors.New("unable to find alignment pattern")

type QRExtractor interface {
	Extract(*image.Gray, QRLocation)
}

type QRExtract struct{}

func (QRExtract) Extract(prepared *image.Gray, loc QRLocation) (QRData, error) {
	size := 17 + loc.Version*4
	p, err := determinePerspective(prepared, loc.Version, size, loc)
	if err != nil {
		return QRData{}, err
	}

	start := loc.TopLeft.Sub(p.dy.Mul(3)).Sub(p.ddy.Mul(3))

	data := make([]byte, 0, size*size)

	dx := p.dx.Sub(p.ddx.Mul(3))
	dy := p.dy.Sub(p.ddy.Mul(3))

	for _i := uint32(0); _i < size; _i++ {
		line := start.Sub(dx.Mul(3))

		for _i := uint32(0); _i < size; _i++ {
			x, y := line.X, line.Y
			pixel := prepared.GrayAt(int(math.Round(x)), int(math.Round(y))).Y
			data = append(data, pixel)
			line = line.Add(dx)
		}
		dx = dx.Add(p.ddx)
		start = start.Add(dy)
		dy = dy.Add(p.ddy)
	}

	return QRData{Data: data, Version: loc.Version, Side: 4*loc.Version + 17}, nil
}

// size := 17 + loc.Version*4

type perspective struct {
	dx, ddx, dy, ddy Point
}

func determinePerspective(
	prepared *image.Gray,
	version, size uint32,
	loc QRLocation,
) (perspective, error) {

	dx := loc.TopRight.Sub(loc.TopLeft)
	dx = dx.Div(float64(size) - 7)

	dy := loc.BottomLeft.Sub(loc.TopLeft)
	dy = dy.Div(float64(size) - 7)

	if version == 1 {
		return perspective{
			dx: dx,
			dy: dy,
		}, nil
	}

	estAlignment := Point{
		X: loc.TopRight.Sub(dx.Mul(3)).Add(dy.Mul(float64(size - 10))).X,
		Y: loc.BottomLeft.Add(dx.Mul(float64(size - 10))).Sub(dy.Mul(3)).Y,
	}

	found := false

distLoop:
	for i := 0; i < 4; i++ {
	scaleLoop:
		for _, j := range []float64{0, 1, -1, 2, -2, 3} {
			scale := 1 + j/10

			if i == 0 {
				if isAlignment(prepared, estAlignment, dx, dy, scale) {
					found = true
					break distLoop
				}
				continue scaleLoop
			}

			for x := -i; x <= i; x++ {
				alignment := estAlignment.Add(dx.Mul(float64(x) / 2).Sub(dy.Mul(float64(i) / 2)))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}

				alignment = estAlignment.Add(dx.Mul(float64(x) / 2)).Add(dy.Mul(float64(i) / 2))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}
			}

			for y := -i; y < i; y++ {
				alignment := estAlignment.Sub(dx.Mul(float64(i) / 2)).Add(dy.Mul(float64(y) / 2))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}

				alignment = estAlignment.Add(dx.Mul(float64(i) / 2)).Add(dy.Mul(float64(y) / 2))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}
			}
		}
	}

	if !found {
		return perspective{}, errUnableToFindPattern
	}

	var (
		alX    = uint32(math.Round(estAlignment.X))
		alY    = uint32(math.Round(estAlignment.Y))
		leftX  = uint32(0)
		rightX = uint32(prepared.Rect.Dx())
	)

	xFunc := revIndex(alX)
	for x := uint32(0); x < alX; x++ {
		x := xFunc(x)
		if prepared.GrayAt(int(x), int(alY)).Y == 255 {
			leftX = x
			break
		}
	}

	for x := alX; x < uint32(prepared.Rect.Dx()); x++ {
		if prepared.GrayAt(int(x), int(alY)).Y == 255 {
			rightX = x
			break
		}
	}

	estAlignment.X = float64(leftX+rightX) / 2

	alX = uint32(math.Round(estAlignment.X))
	alY = uint32(math.Round(estAlignment.Y))
	var (
		topY    uint32
		bottomY = uint32(prepared.Rect.Dy())
	)

	yFunc := revIndex(alY)
	for y := uint32(0); y < alY; y++ {
		y := yFunc(y)
		if prepared.GrayAt(int(alX), int(y)).Y == 255 {
			topY = y
			break
		}
	}

	for y := alY; y < uint32(prepared.Rect.Dy()); y++ {
		if prepared.GrayAt(int(alX), int(y)).Y == 255 {
			bottomY = y
			break
		}
	}

	estAlignment.Y = (float64(topY) + float64(bottomY)) / 2

	originEstimate := Point{
		X: loc.TopRight.Sub(dx.Mul(3)).Add(dy.Mul(float64(size - 10))).X,
		Y: loc.BottomLeft.Add(dx.Mul(float64(size - 10))).Sub(dy.Mul(3)).Y,
	}

	delta := estAlignment.Sub(originEstimate)

	delta = delta.Div(float64((size - 10) * (size - 10)))

	return perspective{dx, delta, dy, Point{}}, nil
}

func isAlignment(prepared *image.Gray, p, dx, dy Point, scale float64) bool {
	if p.X < 0 || p.Y < 0 {
		return false
	}

	dx = dx.Mul(scale)
	dy = dy.Mul(scale)

	topLeft := p.Sub(dx.Mul(2)).Sub(dy.Mul(2))
	if topLeft.X < 0 || topLeft.Y < 0 {
		return false
	}

	bottomRight := p.Add(dx.Mul(2)).Add(dy.Mul(2))
	if bottomRight.X > float64(prepared.Rect.Dx()) ||
		bottomRight.Y > float64(prepared.Rect.Dy()) {
		return false
	}

	for x := -2; x < 2; x++ {
		twiceUp := p.Sub(dx.Mul(float64(x))).Sub(dy.Mul(2))
		if prepared.GrayAt(int(math.Round(twiceUp.X)), int(math.Round(twiceUp.Y))).Y == 255 {
			return false
		}

		twiceDown := p.Sub(dx.Mul(float64(x))).Add(dy.Mul(2))
		if prepared.GrayAt(int(math.Round(twiceDown.X)), int(math.Round(twiceDown.Y))).Y == 255 {
			return false
		}
	}

	for y := -1; y < 1; y++ {
		twiceLeft := p.Sub(dx.Mul(2)).Sub(dy.Mul(float64(y)))
		if prepared.GrayAt(int(math.Round(twiceLeft.X)), int(math.Round(twiceLeft.Y))).Y == 255 {
			return false
		}

		twiceRight := p.Add(dx.Mul(2)).Sub(dy.Mul(float64(y)))
		if prepared.GrayAt(int(math.Round(twiceLeft.X)), int(math.Round(twiceRight.Y))).Y == 255 {
			return false
		}

		left := p.Sub(dx).Sub(dy.Mul(float64(y)))
		if prepared.GrayAt(int(math.Round(left.X)), int(math.Round(left.Y))).Y == 0 {
			return false
		}

		// different from the original source code, might be wrong?
		right := p.Add(dx).Sub(dy.Mul(float64(y)))
		if prepared.GrayAt(int(math.Round(right.X)), int(math.Round(right.Y))).Y == 0 {
			return false
		}
	}

	up := p.Sub(dy)
	if prepared.GrayAt(int(math.Round(up.X)), int(math.Round(up.Y))).Y == 0 {
		return false
	}

	down := p.Add(dy)
	if prepared.GrayAt(int(math.Round(down.X)), int(math.Round(down.Y))).Y == 0 {
		return false
	}

	return prepared.GrayAt(int(math.Round(p.X)), int(math.Round(p.Y))).Y == 0
}
