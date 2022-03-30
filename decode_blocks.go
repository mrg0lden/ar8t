package ar8t

import "fmt"

type codewords struct {
	currentByte, bitCount byte
	blocks                blocks
}

func (c *codewords) addBit(bit byte) {
	c.currentByte *= 2 //shifts to the left by one
	c.currentByte += bit
	c.bitCount += 1

	if c.bitCount == 8 {
		c.blocks.push(c.currentByte)
		c.currentByte = 0
		c.bitCount = 0
	}
}

type blocks struct {
	blockInfo    []BlockInfo
	blocks       [][]byte
	round        uint
	maxDataRound uint
	block        uint
	dataBlocks   bool
}

func newBlocks(blockInfo []BlockInfo) blocks {
	blocksSlice := [][]byte{}
	maxDataRound := uint(0)

	for _, info := range blockInfo {
		if uint(info.DataPer) > maxDataRound {
			maxDataRound = uint(info.DataPer)
		}
		blocksSlice = append(blocksSlice, []byte{})
	}

	return blocks{
		blockInfo:    blockInfo,
		blocks:       blocksSlice,
		maxDataRound: maxDataRound,
		dataBlocks:   true,
	}
}

func (bl *blocks) push(b byte) {
	for bl.dataBlocks &&
		bl.round > uint(bl.blockInfo[bl.block].DataPer)-1 {
		bl.increaseCount()
	}

	bl.blocks[bl.block] = append(bl.blocks[bl.block], b)
	bl.increaseCount()
}

func (bl *blocks) increaseCount() {
	if bl.block != uint(len(bl.blockInfo)-1) {
		bl.block++
		return
	}

	bl.block = 0
	bl.round += 1

	if bl.round == bl.maxDataRound {
		bl.dataBlocks = false
	}
}

type alignmentLocation struct {
	start, step uint32
}

type QRMask = func(QRData, uint32, uint32) byte

func Blocks(data QRData, level ECLevel, mask QRMask) ([][]byte, error) {
	blockInfo, err := GetBlockInfo(int(data.Version), level)
	if err != nil {
		return nil, err
	}

	cwords := codewords{blocks: newBlocks(blockInfo)}
	x := data.Side - 1

	loc, err := getAlignmentLocation(data.Version)
	if err != nil {
		return nil, err
	}

	for {
		rangeY := yRange(x, data.Side)

		for rangeY.Next() {
			y := rangeY.Value()
			if isData(data, loc, x, y) {
				cwords.addBit(mask(data, x, y))
			}

			if isData(data, loc, x-1, y) {
				cwords.addBit(mask(data, x, y))
			}
		}

		if x == 1 {
			break
		}

		x -= 2
		if x == 6 {
			// skip timing pattern
			x = 5
		}
	}

	blockInfo, err = GetBlockInfo(int(data.Version), level)
	blocks := cwords.blocks.blocks

	if len(blocks) != len(blockInfo) {
		return nil, fmt.Errorf("expected %d blocks but found %d",
			len(blockInfo), len(blocks))
	}

	for i := 0; i < len(blocks); i++ {
		if int(blockInfo[i].TotalPer) != len(blocks[i]) {
			return nil, fmt.Errorf("expected %d codewords in block %d but found %d",
				blockInfo[i].TotalPer, i, len(blocks[i]))
		}
	}

	return blocks, nil
}

func yRange(x, side uint32) Iterator[uint32] {
	if x < 6 {
		x++
	}

	if (int64(x)-int64(side)+1)%4 == 0 {
		return _range[uint32]{0, side}.Reverse().Iter()
	}
	return _range[uint32]{0, side}.Iter()
}

func isData(data QRData, loc alignmentLocation, x, y uint32) bool {
	// copied as is TBH

	// timing patterns
	if x == 6 || y == 6 {
		return false
	}

	// top left locator pattern
	if x < 9 && y < 9 {
		return false
	}

	// top right locator pattern
	if x > data.Side-9 && y < 9 {
		return false
	}

	// bottom left locator pattern
	if x < 9 && y > data.Side-9 {
		return false
	}

	// top right version info
	if data.Version >= 7 && x > data.Side-12 && y < 6 {
		return false
	}

	// buttom left version info
	if data.Version >= 7 && y > data.Side-12 && x < 6 {
		return false
	}

	if x == data.Side-9 && y < 9 {
		return true
	}

	if y == data.Side-9 && x < 9 {
		return true
	}

	if isAlignmentCoord(loc, x) && isAlignmentCoord(loc, y) {
		return false
	}

	return true
}

func isAlignmentCoord(loc alignmentLocation, coord uint32) bool {
	switch {
	case coord >= 4 && (coord-4)%6 <= 4:
		return true
	case coord < loc.start-2:
		return false
	case (coord-(loc.start-2))%loc.step <= 4:
		return true
	default:
		return false
	}
}

func getAlignmentLocation(version uint32) (alignmentLocation, error) {
	switch version {
	// no alignment patterns for version 1 but this saves some exception paths
	case 1:
		return alignmentLocation{0, 1000}, nil
	// only one aligalignmentLocation{nment pattern for versions 2-6 but this saves some exception path}s
	case 2:
		return alignmentLocation{8, 1000}, nil
	case 3:
		return alignmentLocation{2, 1000}, nil
	case 4:
		return alignmentLocation{6, 1000}, nil
	case 5:
		return alignmentLocation{0, 1000}, nil
	case 6:
		return alignmentLocation{4, 1000}, nil
	// multiple aligalignmentLocation{nment pattern}s
	case 7:
		return alignmentLocation{2, 16}, nil
	case 8:
		return alignmentLocation{4, 18}, nil
	case 9:
		return alignmentLocation{6, 20}, nil
	case 10:
		return alignmentLocation{8, 22}, nil
	case 11:
		return alignmentLocation{0, 24}, nil
	case 12:
		return alignmentLocation{2, 26}, nil
	case 13:
		return alignmentLocation{4, 28}, nil
	case 14:
		return alignmentLocation{6, 20}, nil
	case 15:
		return alignmentLocation{6, 22}, nil
	case 16:
		return alignmentLocation{6, 24}, nil
	case 17:
		return alignmentLocation{0, 24}, nil
	case 18:
		return alignmentLocation{0, 26}, nil
	case 19:
		return alignmentLocation{0, 28}, nil
	case 20:
		return alignmentLocation{4, 28}, nil
	case 25:
		return alignmentLocation{2, 26}, nil
	case 36:
		return alignmentLocation{4, 26}, nil
	case 40:
		return alignmentLocation{0, 28}, nil
	default:
		return alignmentLocation{}, errUnsupportedVersion
	}
}
