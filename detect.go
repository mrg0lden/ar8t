package ar8t

import (
	"image"
	"math"
)

type Detector interface {
	Detect(prepared *image.Gray) []QRLocation
}

/// Scan a prepared image for QR Codes
///
/// The general idea of this method is as follows:
/// 1. Scan line by line horizontally for possible QR Finder patterns (the three squares)
/// 2. If a possible pattern is found, check vertically and diagonally to confirm it is indeed a pattern
/// 3. Try to find combinations of three patterns that are perpendicular and with similar distance that form a complete QR Code
var _ Detector = LineScan{}

type LineScan struct{}

type QRFinderPosition struct {
	Location                   Point
	ModuleSize, LastModuleSize float64
}

type refineFunc = func(*image.Gray, Point, float64) (QRFinderPosition, bool)

func (s LineScan) Detect(prepared *image.Gray) []QRLocation {
	// The order of refinement is important.
	// The candidate is found in horizontal direction, so the first refinement is vertical
	refineFuncs := []struct {
		refineFunc
		dx, dy     float64
		isDiagonal bool
	}{
		{refineVertical, 0, 1, false},
		{refineHorizontal, 1, 0, false},
		{refineDiagonal, 1, 1, true},
	}

	candidates := []QRFinderPosition{}
	lastPixel := uint8(127)
	pattern := QRFinderPattern{}

	for y := 0; y < prepared.Rect.Dy(); y++ {
	pixels:
		for x := 0; x < prepared.Rect.Dx(); x++ {
			p := prepared.GrayAt(x, y).Y
			// a new line, make a new QRFinderPattern
			if x == 0 {
				lastPixel = 127
				pattern = QRFinderPattern{}
			}

			if p == lastPixel {
				pattern[6] += 1

				if x != prepared.Rect.Dx()-1 {
					continue
				}
			}

			if !pattern.LooksLikeFinder() {
				lastPixel = p
				pattern.Slide()
				continue
			}

			moduleSize := pattern.EstimateModuleSize()

			finder := Point{
				X: float64(x) - moduleSize*3.5,
				Y: float64(y),
			}

			for _, candidate := range candidates {
				if distance(finder, candidate.Location) < 7*moduleSize {
					lastPixel = p
					pattern.Slide()
					continue pixels
				}
			}

			for _, refineFunc := range refineFuncs {
				vert, ok := refineFunc.refineFunc(prepared, finder, moduleSize)

				if !ok {
					lastPixel = p
					pattern.Slide()

					continue pixels
				}

				if !refineFunc.isDiagonal {
					halfFinder := vert.LastModuleSize * 3.5
					finder.X = vert.Location.X - refineFunc.dx*halfFinder
					finder.Y = vert.Location.Y - refineFunc.dy*halfFinder
					moduleSize = vert.ModuleSize
				}
			}

			candidates = append(candidates, QRFinderPosition{
				Location:       finder,
				ModuleSize:     moduleSize,
				LastModuleSize: 0,
			})

			lastPixel = p
			pattern.Slide()

		}
	}

	locations := []QRLocation{}

	maxCandidates := len(candidates)

	for candidate1 := 0; candidate1 < maxCandidates; candidate1++ {
		for candidate2 := candidate1 + 1; candidate2 < maxCandidates; candidate2++ {
			diff1 := diff(
				candidates[candidate1].ModuleSize,
				candidates[candidate2].ModuleSize,
			)

			if diff1 > 0.1 {
				continue
			}

			for candidate3 := candidate2 + 1; candidate3 < maxCandidates; candidate3++ {
				diff2 := diff(
					candidates[candidate1].ModuleSize,
					candidates[candidate3].ModuleSize,
				)

				if diff2 > 0.1 {
					continue
				}

				if loc, ok := findQR(
					[...]Point{
						candidates[candidate1].Location,
						candidates[candidate2].Location,
						candidates[candidate3].Location,
					},
					candidates[candidate1].ModuleSize,
				); ok {
					locations = append(locations, loc)
				}
			}

		}
	}

	return locations

}

func refineHorizontal(prepared *image.Gray, finder Point, moduleSize float64) (QRFinderPosition, bool) {
	startX := refineCalcStart(finder.X, moduleSize)
	endX := refineCalcEnd(finder.X, uint32(prepared.Rect.Dx()), moduleSize)

	y := uint32(math.Round(finder.Y))

	return refine(prepared, moduleSize, startX, endX, y, y, false)

}

func refineVertical(prepared *image.Gray, finder Point, moduleSize float64) (QRFinderPosition, bool) {
	startY := refineCalcStart(finder.Y, moduleSize)
	endY := refineCalcEnd(finder.Y, uint32(prepared.Rect.Dy()), moduleSize)

	x := uint32(math.Round(finder.X))

	return refine(prepared, moduleSize, x, x, startY, endY, false)
}

func refineDiagonal(prepared *image.Gray, finder Point, moduleSize float64) (QRFinderPosition, bool) {
	side := 5 * moduleSize
	var startX, startY float64

	switch {
	case finder.X < side && finder.Y < side:
		if finder.X < finder.Y {
			startY = finder.Y - finder.X
			break
		}
		startX = finder.X - finder.Y
	case finder.X < side:
		startY = finder.Y - finder.X
	case finder.Y < side:
		startX = finder.X - finder.Y
	default:
		startX = finder.X - side
		startY = finder.Y - side
	}

	rangeEndCalc := func(v, d float64) uint32 {
		return uint32(math.Round(min(v+side, d)))
	}

	xMin, xMax := uint32(math.Round(startX)),
		rangeEndCalc(finder.X, float64(prepared.Rect.Dx()))

	yMin, yMax := uint32(math.Round(startY)),
		rangeEndCalc(finder.Y, float64(prepared.Rect.Dy()))

	return refine(prepared, moduleSize, xMin, xMax, yMin, yMax, true)

}

func refine(prepared *image.Gray, moduleSize float64, xMin, xMax, yMin, yMax uint32, isDiagonal bool) (QRFinderPosition, bool) {
	var (
		lastPixel    uint8 = 127
		pattern            = QRFinderPattern{}
		lastX, lastY uint32
	)

	x, y := xMin, yMin

	for {

		p := prepared.GrayAt(int(x), int(y)).Y

		switch {
		case p == lastPixel:
			pattern[6]++
		case pattern.LooksLikeFinder() &&
			(diff(moduleSize, pattern.EstimateModuleSize()) < 0.2 || isDiagonal):
			newEstModSize := (moduleSize + pattern.EstimateModuleSize()) / 2
			return QRFinderPosition{
				Location: Point{
					X: float64(x),
					Y: float64(y),
				},
				ModuleSize:     newEstModSize,
				LastModuleSize: pattern.EstimateModuleSize(),
			}, true
		default:
			lastPixel = p
			pattern.Slide()
		}

		lastX, lastY = x, y

		switch {
		case isDiagonal && x < xMax-1 && y < yMax-1:
			x++
			y++
			continue
		case !isDiagonal && xMin == xMax && y < yMax-1:
			y++
			continue
		case !isDiagonal && yMin == yMax && x < xMax-1:
			x++
			continue
		}

		break
	}

	if pattern.LooksLikeFinder() &&
		(diff(moduleSize, pattern.EstimateModuleSize()) < 0.2 || isDiagonal) {
		newEstModSize := (moduleSize + pattern.EstimateModuleSize()) / 2
		return QRFinderPosition{
			Location: Point{
				X: float64(lastX),
				Y: float64(lastY),
			},
			ModuleSize:     newEstModSize,
			LastModuleSize: pattern.EstimateModuleSize(),
		}, true
	}

	return QRFinderPosition{}, false
}

type QRFinderPattern [7]uint

func (p *QRFinderPattern) Slide() {
	if float64(p[6]) < float64(p[5])/10 && p[4] != 0 {
		p[6] += p[5]
		p[5] = p[4]
		p[4] = p[3]
		p[3] = p[2]
		p[2] = p[1]
		p[1] = p[0]
		p[0] = 0
		return
	}

	p[0] = p[1]
	p[1] = p[2]
	p[2] = p[3]
	p[3] = p[4]
	p[4] = p[5]
	p[5] = p[6]
	p[6] = 1
}

func (p *QRFinderPattern) EstimateModuleSize() float64 {
	return float64(p[2]+p[3]+p[4]+p[5]+p[6]) / 7
}

func (p *QRFinderPattern) LooksLikeFinder() bool {
	totalSize := p[2] + p[3] + p[4] + p[5] + p[6]

	if totalSize < 7 {
		return false
	}

	moduleSize := float64(totalSize) / 7
	maxVariance := moduleSize / 1.5

	check := func(v ...uint) bool {
		for i, val := range v {
			// p[4] special case
			// not good code tbh
			if i == 2 {
				if math.Abs(moduleSize*3-float64(val)) > maxVariance {
					return false
				}

				continue
			}
			if math.Abs(moduleSize-float64(val)) > maxVariance {
				return false
			}
		}
		return true
	}

	return check(p[2], p[3], p[4], p[5], p[6])

}

func diff(a, b float64) float64 {
	if a > b {
		return (a - b) / a
	}
	return (b - a) / b
}

func distance(a, b Point) float64 {
	// square of diff
	dist := float64(a.X-b.X)*float64(a.X-b.X) +
		float64(a.Y-b.Y)*float64(a.Y-b.Y)
	return math.Sqrt(dist)
}

func refineCalcStart(v, moduleSize float64) uint32 {
	return uint32(math.Round(max(v-5*moduleSize, 0)))
}

func refineCalcEnd(v float64, d uint32, moduleSize float64) uint32 {
	return min(
		uint32(math.Round(v+5*moduleSize)),
		d,
	)
}

func findQR(p [3]Point, moduleSize float64) (QRLocation, bool) {
	loc, ok := findQRInternal([...]Point{p[0], p[1], p[2]}, moduleSize)
	if ok {
		return loc, ok
	}

	loc, ok = findQRInternal([...]Point{p[1], p[0], p[2]}, moduleSize)
	if ok {
		return loc, ok
	}

	loc, ok = findQRInternal([...]Point{p[2], p[0], p[1]}, moduleSize)
	if ok {
		return loc, ok
	}

	return QRLocation{}, false
}

func findQRInternal(p [3]Point, moduleSize float64) (QRLocation, bool) {
	var (
		ax = p[1].X - p[0].X
		ay = p[1].Y - p[0].Y
		bx = p[2].X - p[0].X
		by = p[2].Y - p[0].Y
	)

	var (
		crossProduct = -(ax*by - ay*bx)
		lenA         = math.Sqrt(ax*ax + ay*ay)
		lenB         = math.Sqrt(bx*bx + by*by)
	)

	if diff(lenA, lenB) > 0.15 {
		return QRLocation{}, false
	}

	perpendicular := crossProduct / lenA / lenB
	if math.Abs(math.Abs(perpendicular)-1) > 0.05 {
		return QRLocation{}, false
	}

	dist := uint32(math.Round(distance(p[0], p[2])/moduleSize + 7))

	if dist < 20 {
		return QRLocation{}, false
	}

	switch dist % 4 {
	case 0:
		dist += 1
	case 1:
	case 2:
		dist -= 1
	case 3:
		dist -= 2
	default:
		return QRLocation{}, false
	}

	if perpendicular > 0 {
		return QRLocation{
			TopLeft:    p[0],
			TopRight:   p[2],
			BottomLeft: p[1],
			ModuleSize: moduleSize,
			Version:    (dist - 17) / 4,
		}, true
	}

	return QRLocation{
		TopLeft:    p[0],
		TopRight:   p[1],
		BottomLeft: p[2],
		ModuleSize: moduleSize,
		Version:    (dist - 17) / 4,
	}, true
}
