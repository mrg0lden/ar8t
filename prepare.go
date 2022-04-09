package ar8t

/// Reduce the image to black/white by calculating local thresholds
///
/// The algorithm runs the following steps:
/// 1. Divide the image into blocks and count the cumulative grayscale value of all pixels in the block
/// 2. For each block of blocks, take mean grayscale value by adding each block's value and dividing by total number of pixels
/// 3. For each pixel in the image, see if the grayscale value of that pixel exceeds the mean of its corresponding block.
///    If so, output a white pixel. If not, output a black pixel

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"golang.org/x/exp/constraints"
)

type BlockSize uint32

type stats struct {
	total, count uint64
	mean         float64
}

type BlockedMean struct {
	blockSize     BlockSize
	blockMeanSize BlockSize
}

func NewBlockedMean(blockSize, blockMeanSize uint32) BlockedMean {
	return BlockedMean{BlockSize(blockSize), BlockSize(blockMeanSize)}
}

func (b BlockedMean) Prepare(img image.Image) *image.Gray {
	gray := imaging.Grayscale(img)

	blockMap := b.asBlockMap(gray)
	blockMeanMap := b.toBlockMeanMap(blockMap, gray.Rect)

	return b.toThreshold(gray, blockMeanMap)
}

func (b BlockedMean) asBlockMap(gray *image.NRGBA) []stats {
	blockWidth, blockHeight := asBlockCoords(gray.Rect.Dx(), gray.Rect.Dy(), b.blockSize)

	blocks := make([]stats, (blockWidth+1)*(blockHeight+1))

	for y := 0; y < gray.Rect.Dy(); y++ {
		for x := 0; x < gray.Rect.Dx(); x++ {
			p := gray.NRGBAAt(x, y)
			coordX, coordY := asBlockCoords(x, y, b.blockSize)
			stat := &blocks[toIndex(coordX, coordY, blockWidth)]

			stat.total += uint64(p.R)
			stat.count += 1
		}
	}

	for i, stat := range blocks {
		blocks[i].mean = float64(stat.total) / float64(stat.count)
	}

	return blocks
}

func (b BlockedMean) toBlockMeanMap(blocks []stats, rect image.Rectangle) []stats {
	blockStride := BlockCoord((b.blockMeanSize - 1) / 2)
	blockW, blockH := asBlockCoords(rect.Dx(), rect.Dy(), b.blockSize)

	blockMeans := make([]stats, (blockW+1)*(blockH+1))

	var blockX, blockY BlockCoord
	// range_include
	for blockX = 0; blockX <= blockW; blockX++ {
		for blockY = 0; blockY <= blockH; blockY++ {
			xStart := max(0, blockX.saturatingSub(blockStride))
			xEnd := min(blockW, blockX+blockStride)
			yStart := max(0, blockY.saturatingSub(blockStride))
			yEnd := min(blockH, blockY+blockStride)

			var total, count uint64

			for x := xStart; x < xEnd; x++ {
				for y := yStart; y < yEnd; y++ {
					// Original author[piderman314]:
					// Make sure to take the pixel counts from each of the blocks directly
					// Because the size of the image does not have to be an exact multiple of the size in blocks,
					// some blocks can have differing pixel counts
					stat := blocks[toIndex(x, y, blockW)]
					total += stat.total
					count += stat.count
				}
			}

			blockMeans[toIndex(blockX, blockY, blockW)].mean = float64(total) / float64(count)

		}
	}

	return blockMeans

}

func (b BlockedMean) toThreshold(gray *image.NRGBA, blockMeans []stats) *image.Gray {
	actualGray := image.NewGray(gray.Rect)
	for y := 0; y < gray.Rect.Dy(); y++ {
		for x := 0; x < gray.Rect.Dx(); x++ {
			p := gray.NRGBAAt(x, y)

			blockWidth, _ := asBlockCoords(gray.Rect.Dx(), gray.Rect.Dy(), b.blockSize)
			coordX, coordY := asBlockCoords(x, y, b.blockSize)

			mean := blockMeans[toIndex(coordX, coordY, blockWidth)].mean
			grayColor := color.Gray{}
			switch {
			case mean > 250:
				grayColor.Y = 255
			case mean < 5:
				// do nothing
			case float64(p.R) > mean:
				grayColor.Y = 255
			}

			actualGray.SetGray(x, y, grayColor)

		}
	}

	return actualGray
}

func asBlockCoords(w, h int, s BlockSize) (x, y BlockCoord) {
	x = BlockCoord(w) / BlockCoord(s)
	y = BlockCoord(h) / BlockCoord(s)
	return
}

func toIndex(coordX, coordY, w BlockCoord) uint32 {
	return uint32(coordY*(w+1) + coordX)
}

type BlockCoord uint32

func (c BlockCoord) saturatingSub(other BlockCoord) (res BlockCoord) {
	return c - min(c, other)
}

func max[T constraints.Ordered](a, b T) T {
	if a >= b {
		return a
	}

	return b
}

func min[T constraints.Ordered](a, b T) T {
	if a <= b {
		return a
	}
	return b
}
