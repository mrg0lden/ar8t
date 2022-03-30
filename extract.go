package ar8t

import (
	"errors"
	"image"
)

var errUnableToFindPattern = errors.New("unable to find alignment pattern")

type QRExtractor interface {
	Extract(*image.Gray, QRLocation)
}

type QRExtract struct{}

func (QRExtract) Extract(prepared *image.Gray, loc QRLocation) (QRData, error)

// size := 17 + loc.Version*4

type perspective struct {
	dx, ddx, dy, ddy image.Point
}

func determinePerspective(
	prepared *image.Gray,
	version, size uint32,
	loc QRLocation,
) (perspective, error) {

	dx := loc.TopRight.Sub(loc.TopLeft)
	dx = dx.Div(int(float64(size) / 7))

	dy := loc.BottomLeft.Sub(loc.TopLeft)
	dy = dy.Div(int(float64(size) / 7))

	if version == 1 {
		return perspective{
			dx: dx,
			dy: dy,
		}, nil
	}

	estAlignment := image.Point{
		X: loc.TopRight.Sub(dx.Mul(3)).Add(dy.Mul(int(size - 10))).X,
		Y: loc.BottomLeft.Add(dx.Mul(int(size - 10))).Sub(dy.Mul(3)).Y,
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
				alignment := estAlignment.Add(dx.Mul(x / 2).Sub(dy.Mul(i / 2)))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}

				alignment = estAlignment.Add(dx.Mul(x / 2)).Add(dy.Mul(i / 2))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}
			}

			for y := -i + 1; y < i; y++ {
				alignment := estAlignment.Sub(dx.Mul(i / 2)).Add(dy.Mul(y / 2))
				if isAlignment(prepared, alignment, dx, dy, scale) {
					estAlignment = alignment
					found = true
					break distLoop
				}

				alignment = estAlignment.Add(dx.Mul(i / 2)).Add(dy.Mul(y / 2))
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
		alX    = estAlignment.X
		alY    = estAlignment.Y
		leftX  = 0
		rightX = prepared.Rect.Dx()
	)

	for x := alX; x > 0; x-- {
		if prepared.GrayAt(x, alY).Y == 255 {
			leftX = x
			break
		}
	}

	for x := alX; x < prepared.Rect.Dx(); x++ {
		if prepared.GrayAt(x, alY).Y == 255 {
			rightX = x
			break
		}
	}

	estAlignment.X = (leftX + rightX) / 2

	alX = estAlignment.X
	alY = estAlignment.Y
	var (
		topY    = 0
		bottomY = 0
	)

	for y := alY; y > 0; y-- {
		if prepared.GrayAt(alX, y).Y == 255 {
			topY = y
			break
		}
	}

	for y := alY; y < prepared.Rect.Dy(); y++ {
		if prepared.GrayAt(alX, y).Y == 255 {
			bottomY = y
			break
		}
	}

	estAlignment.Y = (topY + bottomY) / 2

	originEstimate := image.Point{
		X: loc.TopRight.Sub(dx.Mul(3)).Add(dy.Mul(int(size - 10))).X,
		Y: loc.BottomLeft.Add(dx.Mul(int(size - 10))).Sub(dy.Mul(3)).Y,
	}

	delta := estAlignment.Sub(originEstimate)

	delta = delta.Div(int(size-10) * int(size-10))

	return perspective{dx, delta, dy, image.Point{}}, nil
}

func isAlignment(prepared *image.Gray, p, dx, dy image.Point, scale float64) bool {
	if p.X < 0 || p.Y < 0 {
		return false
	}

	dx = dx.Mul(int(scale))
	dy = dy.Mul(int(scale))

	topLeft := p.Sub(dx.Mul(2)).Sub(dy.Mul(2))
	if topLeft.X < 0 || topLeft.Y < 0 {
		return false
	}

	bottomRight := p.Add(dx.Mul(2)).Add(dx.Mul(2))
	if bottomRight.X > prepared.Rect.Dx() || bottomRight.Y > prepared.Rect.Dy() {
		return false
	}

	for x := -2; x < 2; x++ {
		twiceUp := p.Sub(dx.Mul(x)).Sub(dy.Mul(2))
		if prepared.GrayAt(twiceUp.X, twiceUp.Y).Y == 255 {
			return false
		}

		twiceDown := p.Sub(dx.Mul(x)).Add(dy.Mul(2))
		if prepared.GrayAt(twiceDown.X, twiceDown.Y).Y == 255 {
			return false
		}
	}

	for y := -1; y < 1; y++ {
		twiceLeft := p.Sub(dx.Mul(2)).Sub(dy.Mul(y))
		if prepared.GrayAt(twiceLeft.X, twiceLeft.Y).Y == 255 {
			return false
		}

		twiceRight := p.Add(dx.Mul(2)).Sub(dy.Mul(y))
		if prepared.GrayAt(twiceLeft.X, twiceRight.Y).Y == 255 {
			return false
		}
	}

	up := p.Sub(dy)
	if prepared.GrayAt(up.X, up.Y).Y == 255 {
		return false
	}

	down := p.Add(dy)
	if prepared.GrayAt(down.X, down.Y).Y == 255 {
		return false
	}

	return prepared.GrayAt(p.X, p.Y).Y == 0
}
