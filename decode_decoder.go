package ar8t

//QRDecoder is ready to use as is
type QRDecoder struct{}

func (QRDecoder) Decode(qrData QRData) ([]byte, error) {
	ecLevel, mask, err := Format(qrData)
	if err != nil {
		return nil, err
	}

	blocks, err := Blocks(qrData, ecLevel, mask)
	if err != nil {
		return nil, err
	}

	blockInfo, err := GetBlockInfo(qrData.Version, ecLevel)
	if err != nil {
		return nil, err
	}

	allBlocks := []byte{}

	for i := 0; i < len(blocks) && i < len(blockInfo); i++ {
		corrected, err := Correct(blocks[i], blockInfo[i])
		if err != nil {
			return nil, err
		}

		allBlocks = append(allBlocks, corrected[:blockInfo[i].DataPer]...)
	}

	data, err := Data(allBlocks, qrData.Version)
	if err != nil {
		return nil, err
	}

	return data, nil

}

// type QRDecoderWithInfo struct{}

// func (QRDecoderWithInfo) Decode(qrData QRData)
