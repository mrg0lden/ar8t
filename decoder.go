package ar8t

import (
	"errors"
	"image"
)

type DefaultDecoder struct{}

var ErrNoSymbolsFound = errors.New("no symbols found")

func (DefaultDecoder) Decode(src image.Image) ([][]byte, error) {
	prepared := NewBlockedMean(3, 7).Prepare(src)
	locations := LineScan{}.Detect(prepared)

	if len(locations) == 0 {
		return nil, ErrNoSymbolsFound
	}

	allDecoded := [][]byte{}
	// errs := []error{}

	for _, location := range locations {
		extracted, err := QRExtract{}.Extract(prepared, location)
		if err != nil {
			continue
		}
		decoded, err := QRDecoder{}.Decode(extracted)
		if err != nil {
			continue
			// errs = append(errs, err)
		}

		allDecoded = append(allDecoded, decoded)

	}

	// for debug mode

	// errStr := strings.Builder{}

	// for _, err := range errs {
	// 	errStr.WriteString("err: ")
	// 	errStr.WriteString(err.Error())
	// 	errStr.WriteByte('\n')
	// }

	return allDecoded, nil
}
