package ar8t

import (
	"errors"
)

type BlockInfo struct {
	BlockCount        byte
	TotalPer, DataPer byte
	EC_Cap            byte
}

var (
	errUnsupportedVersion = errors.New("ar8t: unsupported version")
	errInvalidECLevel     = errors.New("ar8t: invalid error correction level")
)

func GetBlockInfo(version uint32, level ECLevel) ([]BlockInfo, error) {
	if version < 1 || version > 40 {
		return nil, errUnsupportedVersion
	}

	if level < ECLevelLow || level > ECLevelHigh {
		return nil, errInvalidECLevel
	}

	infoExpanded := []BlockInfo{}

	blockInfo := levelsBlockInfo[version][level]

	for _, bi := range blockInfo {
		for i := byte(0); i < bi.BlockCount; i++ {
			infoExpanded = append(infoExpanded, bi)
		}
	}

	return infoExpanded, nil
}

var levelsBlockInfo = map[uint32]map[ECLevel][]BlockInfo{
	1: {
		ECLevelLow:      {{1, 26, 19, 2}},
		ECLevelMedium:   {{1, 26, 16, 4}},
		ECLevelQuartile: {{1, 26, 13, 6}},
		ECLevelHigh:     {{1, 26, 9, 8}},
	},

	2: {
		ECLevelLow:      {{1, 44, 34, 4}},
		ECLevelMedium:   {{1, 44, 28, 8}},
		ECLevelQuartile: {{1, 44, 22, 11}},
		ECLevelHigh:     {{1, 44, 16, 14}},
	},

	3: {
		ECLevelLow:      {{1, 70, 55, 7}},
		ECLevelMedium:   {{1, 70, 44, 13}},
		ECLevelQuartile: {{2, 35, 17, 9}},
		ECLevelHigh:     {{2, 35, 13, 11}},
	},

	4: {
		ECLevelLow:      {{1, 100, 80, 10}},
		ECLevelMedium:   {{2, 50, 32, 9}},
		ECLevelQuartile: {{2, 50, 24, 13}},
		ECLevelHigh:     {{4, 25, 9, 8}},
	},

	5: {
		ECLevelLow:    {{1, 134, 108, 13}},
		ECLevelMedium: {{2, 67, 43, 12}},
		ECLevelQuartile: {
			{2, 33, 15, 9},
			{2, 34, 16, 9},
		},
		ECLevelHigh: {
			{2, 33, 11, 11},
			{2, 34, 12, 11},
		},
	},

	6: {
		ECLevelLow:      {{2, 86, 68, 9}},
		ECLevelMedium:   {{4, 43, 27, 8}},
		ECLevelQuartile: {{4, 43, 19, 12}},
		ECLevelHigh:     {{4, 43, 15, 14}},
	},

	7: {
		ECLevelLow:    {{2, 98, 78, 10}},
		ECLevelMedium: {{4, 49, 31, 9}},
		ECLevelQuartile: {
			{2, 32, 14, 9},
			{4, 33, 15, 9},
		},
		ECLevelHigh: {
			{4, 39, 13, 13},
			{1, 40, 14, 13},
		},
	},

	8: {
		ECLevelLow: {{2, 121, 97, 12}},
		ECLevelMedium: {
			{2, 60, 38, 11},
			{2, 61, 39, 11},
		},
		ECLevelQuartile: {
			{4, 40, 18, 11},
			{2, 41, 19, 11},
		},
		ECLevelHigh: {
			{4, 40, 14, 13},
			{2, 41, 15, 13},
		},
	},

	9: {
		ECLevelLow: {{2, 146, 116, 15}},
		ECLevelMedium: {
			{3, 58, 36, 11},
			{2, 59, 37, 11},
		},
		ECLevelQuartile: {
			{4, 36, 16, 10},
			{4, 37, 17, 10},
		},
		ECLevelHigh: {
			{4, 36, 12, 12},
			{4, 37, 13, 12},
		},
	},

	10: {
		ECLevelLow: {
			{2, 86, 68, 9},
			{2, 87, 69, 9},
		},
		ECLevelMedium: {
			{4, 69, 43, 13},
			{1, 70, 44, 13},
		},
		ECLevelQuartile: {
			{6, 43, 19, 12},
			{2, 44, 20, 12},
		},
		ECLevelHigh: {
			{6, 43, 15, 14},
			{2, 44, 16, 14},
		},
	},

	11: {
		ECLevelLow: {{4, 101, 81, 10}},
		ECLevelMedium: {
			{1, 80, 50, 15},
			{4, 81, 51, 15},
		},
		ECLevelQuartile: {
			{4, 50, 22, 14},
			{4, 51, 23, 14},
		},
		ECLevelHigh: {
			{3, 36, 12, 12},
			{8, 37, 13, 12},
		},
	},

	12: {
		ECLevelLow: {
			{2, 116, 92, 12},
			{2, 117, 93, 12},
		},
		ECLevelMedium: {
			{6, 58, 36, 11},
			{2, 59, 37, 11},
		},
		ECLevelQuartile: {
			{4, 46, 20, 13},
			{6, 47, 21, 13},
		},
		ECLevelHigh: {
			{7, 42, 14, 14},
			{4, 43, 15, 14},
		},
	},

	13: {
		ECLevelLow: {{4, 133, 107, 13}},
		ECLevelMedium: {
			{8, 59, 37, 11},
			{1, 60, 38, 11},
		},
		ECLevelQuartile: {
			{8, 44, 20, 12},
			{4, 45, 21, 12},
		},
		ECLevelHigh: {
			{12, 33, 11, 11},
			{4, 34, 12, 11},
		},
	},
	14: {
		ECLevelLow: {
			{3, 145, 115, 15},
			{1, 146, 116, 15},
		},
		ECLevelMedium: {
			{4, 64, 40, 12},
			{5, 65, 41, 12},
		},
		ECLevelQuartile: {
			{11, 36, 16, 10},
			{5, 37, 17, 10},
		},
		ECLevelHigh: {
			{11, 36, 12, 12},
			{5, 37, 13, 12},
		},
	},
	15: {
		ECLevelLow: {
			{5, 109, 87, 11},
			{1, 110, 88, 11},
		},
		ECLevelMedium: {
			{5, 65, 41, 12},
			{5, 66, 42, 12},
		},
		ECLevelQuartile: {
			{5, 54, 24, 15},
			{7, 55, 25, 15},
		},
		ECLevelHigh: {
			{11, 36, 12, 12},
			{7, 37, 13, 12},
		},
	},
	16: {
		ECLevelLow: {
			{5, 122, 98, 12},
			{1, 123, 99, 12},
		},
		ECLevelMedium: {
			{7, 73, 45, 14},
			{3, 74, 46, 14},
		},
		ECLevelQuartile: {
			{15, 43, 19, 12},
			{2, 44, 20, 12},
		},
		ECLevelHigh: {
			{3, 45, 15, 15},
			{13, 46, 16, 15},
		},
	},
	17: {
		ECLevelLow: {
			{1, 135, 107, 14},
			{5, 136, 108, 14},
		},
		ECLevelMedium: {
			{10, 74, 46, 14},
			{1, 75, 47, 14},
		},
		ECLevelQuartile: {
			{1, 50, 22, 14},
			{15, 51, 23, 14},
		},
		ECLevelHigh: {
			{2, 42, 14, 14},
			{17, 43, 15, 14},
		},
	},
	18: {
		ECLevelLow: {
			{5, 150, 120, 15},
			{1, 151, 121, 15},
		},
		ECLevelMedium: {
			{9, 69, 43, 13},
			{4, 70, 44, 13},
		},
		ECLevelQuartile: {
			{17, 50, 22, 14},
			{1, 51, 23, 14},
		},
		ECLevelHigh: {
			{2, 42, 14, 14},
			{19, 43, 15, 14},
		},
	},
	19: {
		ECLevelLow: {
			{3, 141, 113, 14},
			{4, 142, 114, 14},
		},
		ECLevelMedium: {
			{3, 70, 44, 13},
			{11, 71, 45, 13},
		},
		ECLevelQuartile: {
			{17, 47, 21, 13},
			{4, 48, 22, 13},
		},
		ECLevelHigh: {
			{9, 39, 13, 13},
			{16, 40, 14, 13},
		},
	},
	20: {
		ECLevelLow: {
			{3, 135, 107, 14},
			{5, 136, 108, 14},
		},
		ECLevelMedium: {
			{3, 67, 41, 13},
			{13, 68, 42, 13},
		},
		ECLevelQuartile: {
			{15, 54, 24, 15},
			{5, 55, 25, 15},
		},
		ECLevelHigh: {
			{15, 43, 15, 14},
			{10, 44, 16, 14},
		},
	},
	21: {
		ECLevelLow: {
			{4, 144, 116, 14},
			{4, 145, 117, 14},
		},
		ECLevelMedium: {{17, 68, 42, 13}},
		ECLevelQuartile: {
			{17, 50, 22, 14},
			{6, 51, 23, 14},
		},
		ECLevelHigh: {
			{19, 46, 16, 15},
			{6, 47, 17, 15},
		},
	},
	22: {
		ECLevelLow: {
			{2, 139, 111, 14},
			{7, 140, 112, 14},
		},
		ECLevelMedium: {{17, 74, 46, 14}},
		ECLevelQuartile: {
			{7, 54, 24, 15},
			{16, 55, 25, 15},
		},
		ECLevelHigh: {{34, 37, 13, 12}},
	},
	23: {
		ECLevelLow: {
			{4, 151, 121, 15},
			{5, 152, 122, 15},
		},
		ECLevelMedium: {
			{4, 75, 47, 14},
			{14, 76, 48, 14},
		},
		ECLevelQuartile: {
			{11, 54, 24, 15},
			{14, 55, 25, 15},
		},
		ECLevelHigh: {
			{16, 45, 15, 15},
			{14, 46, 16, 15},
		},
	},
	24: {
		ECLevelLow: {
			{6, 147, 117, 15},
			{4, 148, 118, 15},
		},
		ECLevelMedium: {
			{6, 73, 45, 14},
			{14, 74, 46, 14},
		},
		ECLevelQuartile: {
			{11, 54, 24, 15},
			{16, 55, 25, 15},
		},
		ECLevelHigh: {
			{30, 46, 16, 15},
			{2, 47, 17, 15},
		},
	},
	25: {
		ECLevelLow: {
			{8, 132, 106, 13},
			{4, 133, 107, 13},
		},
		ECLevelMedium: {
			{8, 75, 47, 14},
			{13, 76, 48, 14},
		},
		ECLevelQuartile: {
			{7, 54, 24, 15},
			{22, 55, 25, 15},
		},
		ECLevelHigh: {
			{22, 45, 15, 15},
			{13, 46, 16, 15},
		},
	},
	26: {
		ECLevelLow: {
			{10, 142, 114, 14},
			{2, 143, 115, 14},
		},
		ECLevelMedium: {
			{19, 74, 46, 14},
			{4, 75, 47, 14},
		},
		ECLevelQuartile: {
			{28, 50, 22, 14},
			{6, 51, 23, 14},
		},
		ECLevelHigh: {
			{33, 46, 16, 15},
			{4, 47, 17, 15},
		},
	},
	27: {
		ECLevelLow: {
			{8, 152, 122, 15},
			{4, 153, 123, 15},
		},
		ECLevelMedium: {
			{22, 73, 45, 14},
			{3, 74, 46, 14},
		},
		ECLevelQuartile: {
			{8, 53, 23, 15},
			{26, 54, 24, 15},
		},
		ECLevelHigh: {
			{12, 45, 15, 15},
			{28, 46, 16, 15},
		},
	},
	28: {
		ECLevelLow: {
			{3, 147, 117, 15},
			{10, 148, 118, 15},
		},
		ECLevelMedium: {
			{3, 73, 45, 14},
			{23, 74, 46, 14},
		},
		ECLevelQuartile: {
			{4, 54, 24, 15},
			{31, 55, 25, 15},
		},
		ECLevelHigh: {
			{11, 45, 15, 15},
			{31, 46, 16, 15},
		},
	},
	29: {
		ECLevelLow: {
			{7, 146, 116, 15},
			{7, 147, 117, 15},
		},
		ECLevelMedium: {
			{21, 73, 45, 14},
			{7, 74, 46, 14},
		},
		ECLevelQuartile: {
			{1, 53, 23, 15},
			{37, 54, 24, 15},
		},
		ECLevelHigh: {
			{19, 45, 15, 15},
			{26, 46, 16, 15},
		},
	},
	30: {
		ECLevelLow: {
			{5, 145, 115, 15},
			{10, 146, 116, 15},
		},
		ECLevelMedium: {
			{19, 75, 47, 14},
			{10, 76, 48, 14},
		},
		ECLevelQuartile: {
			{15, 54, 24, 15},
			{25, 55, 25, 15},
		},
		ECLevelHigh: {
			{23, 45, 15, 15},
			{25, 46, 16, 15},
		},
	},
	31: {
		ECLevelLow: {
			{13, 145, 115, 15},
			{3, 146, 116, 15},
		},
		ECLevelMedium: {
			{2, 74, 46, 14},
			{29, 75, 47, 14},
		},
		ECLevelQuartile: {
			{42, 54, 24, 15},
			{1, 55, 25, 15},
		},
		ECLevelHigh: {
			{23, 45, 15, 15},
			{28, 46, 16, 15},
		},
	},
	32: {
		ECLevelLow: {{17, 145, 115, 15}},
		ECLevelMedium: {
			{10, 74, 46, 14},
			{23, 75, 47, 14},
		},
		ECLevelQuartile: {
			{10, 54, 24, 15},
			{35, 55, 25, 15},
		},
		ECLevelHigh: {
			{19, 45, 15, 15},
			{35, 46, 16, 15},
		},
	},
	33: {
		ECLevelLow: {
			{17, 145, 115, 15},
			{1, 146, 116, 15},
		},
		ECLevelMedium: {
			{14, 74, 46, 14},
			{21, 75, 47, 14},
		},
		ECLevelQuartile: {
			{29, 54, 24, 15},
			{19, 55, 25, 15},
		},
		ECLevelHigh: {
			{11, 45, 15, 15},
			{46, 46, 16, 15},
		},
	},
	34: {
		ECLevelLow: {
			{13, 145, 115, 15},
			{6, 146, 116, 15},
		},
		ECLevelMedium: {
			{14, 74, 46, 14},
			{23, 75, 47, 14},
		},
		ECLevelQuartile: {
			{44, 54, 24, 15},
			{7, 55, 25, 15},
		},
		ECLevelHigh: {
			{59, 46, 16, 15},
			{1, 47, 17, 15},
		},
	},
	35: {
		ECLevelLow: {
			{12, 151, 121, 15},
			{7, 152, 122, 15},
		},
		ECLevelMedium: {
			{12, 75, 47, 14},
			{26, 76, 48, 14},
		},
		ECLevelQuartile: {
			{39, 54, 24, 15},
			{14, 55, 25, 15},
		},
		ECLevelHigh: {
			{22, 45, 15, 15},
			{41, 46, 16, 15},
		},
	},
	36: {
		ECLevelLow: {
			{6, 151, 121, 15},
			{14, 152, 122, 15},
		},
		ECLevelMedium: {
			{6, 75, 47, 14},
			{34, 76, 48, 14},
		},
		ECLevelQuartile: {
			{46, 54, 24, 15},
			{10, 55, 25, 15},
		},
		ECLevelHigh: {
			{2, 45, 15, 15},
			{64, 46, 16, 15},
		},
	},
	37: {
		ECLevelLow: {
			{17, 152, 122, 15},
			{4, 153, 123, 15},
		},
		ECLevelMedium: {
			{29, 74, 46, 14},
			{14, 75, 47, 14},
		},
		ECLevelQuartile: {
			{49, 54, 24, 15},
			{10, 55, 25, 15},
		},
		ECLevelHigh: {
			{24, 45, 15, 15},
			{46, 46, 16, 15},
		},
	},
	38: {
		ECLevelLow: {
			{4, 152, 122, 15},
			{18, 153, 123, 15},
		},
		ECLevelMedium: {
			{13, 74, 46, 14},
			{32, 75, 47, 14},
		},
		ECLevelQuartile: {
			{48, 54, 24, 15},
			{14, 55, 25, 15},
		},
		ECLevelHigh: {
			{42, 45, 15, 15},
			{32, 46, 16, 15},
		},
	},
	39: {
		ECLevelLow: {
			{20, 147, 117, 15},
			{4, 148, 118, 15},
		},
		ECLevelMedium: {
			{40, 75, 47, 14},
			{7, 76, 48, 14},
		},
		ECLevelQuartile: {
			{43, 54, 24, 15},
			{22, 55, 25, 15},
		},
		ECLevelHigh: {
			{10, 45, 15, 15},
			{67, 46, 16, 15},
		},
	},
	40: {
		ECLevelLow: {
			{19, 148, 118, 15},
			{6, 149, 119, 15},
		},
		ECLevelMedium: {
			{18, 75, 47, 14},
			{31, 76, 48, 14},
		},
		ECLevelQuartile: {
			{34, 54, 24, 15},
			{34, 55, 25, 15},
		},
		ECLevelHigh: {
			{20, 45, 15, 15},
			{61, 46, 16, 15},
		},
	},
}
