package ar8t

import (
	"errors"
	"math/bits"

	"golang.org/x/exp/slices"
)

var (
	errFailedToCalcDist  = errors.New("could not calculate error distances")
	errFailedToFixData   = errors.New("error correcting did not fix corrupted data")
	errFailedToCalcSigma = errors.New("could not calculate SIGMA")
)

func Correct(block []byte, blockInfo BlockInfo) ([]byte, error) {
	res, _, err := CorrectWithErrorCount(block, blockInfo)
	return res, err
}

func CorrectWithErrorCount(block []byte, blockInfo BlockInfo) ([]byte, int, error) {
	syndromes, allFine := calculateSyndromes(block, blockInfo)
	if allFine {
		return block, 0, nil
	}

	locs, err := findLocations(blockInfo, syndromes)
	if err != nil {
		return nil, 0, err
	}

	distances, ok := calculateDistances(syndromes, locs)
	if !ok {
		return nil, 0, errFailedToCalcDist
	}

	errCount := 0

	for i := range locs {
		errCount += bits.OnesCount8(distances[i].v)
		index := int(blockInfo.TotalPer) - 1 - locs[i]
		switch {
		case index < 0:
			index = 0
		case index >= int(blockInfo.TotalPer):
			index = int(blockInfo.TotalPer) - 1
		}

		block[index] ^= distances[i].v
	}

	if syndrome(block, exp8[0]) != (gf8{}) {
		return nil, 0, errFailedToFixData
	}

	return block, errCount, nil
}

func calculateSyndromes(block []byte, blockInfo BlockInfo) ([]gf8, bool) {
	syndromes := make([]gf8, blockInfo.EC_Cap*2)

	allFine := true
	for i := uint8(0); i < blockInfo.EC_Cap*2; i++ {
		syndromes[i] = syndrome(block, exp8[i])
		if syndromes[i] != (gf8{}) {
			allFine = false
		}
	}

	return syndromes, allFine
}

func syndrome(block []byte, base gf8) gf8 {
	synd, alpha := gf8{0}, gf8{1}
	iFunc := revIndex(len(block))
	for i := range block {
		i := iFunc(i)
		codeword := block[i]
		synd = synd.AddOrSub(alpha.Mul(gf8{codeword}))
		alpha = alpha.Mul(base)
	}

	return synd
}

func findLocations(info BlockInfo, syndromes []gf8) ([]int, error) {
	z := info.EC_Cap
	eq := make([][]gf8, z)
	for i := uint8(0); i < z; i++ {
		eq[i] = slices.Clone(syndromes)[i : z+i+1]
	}

	sigma, ok := solve(eq, gf8{1}, false)
	if !ok {
		return nil, errFailedToCalcSigma
	}

	locs := []int{}

	for i, exp := range exp8 {
		if uint8(i) > info.TotalPer {
			break
		}

		var (
			x          = exp
			checkValue = sigma[0]
		)

		for _, s := range sigma[1:] {
			checkValue = checkValue.AddOrSub(x.Mul(s))
			x = x.Mul(exp)
		}

		checkValue = checkValue.AddOrSub(x)

		if checkValue == (gf8{}) {
			locs = append(locs, i)
		}
	}

	return locs, nil

}

func calculateDistances(syndromes []gf8, locs []int) ([]gf8, bool) {
	eq := make([][]gf8, len(locs))
	for i := range locs {
		eq[i] = make([]gf8, len(locs)+1)

		for j := range locs {
			eq[i][j] = exp8[(i*locs[j])%255]
		}

		eq[i][len(locs)] = syndromes[i]
	}

	return solve(eq, gf8{1}, false)
}

func solve(eq [][]gf8, one gf8, failOnRank bool) ([]gf8, bool) {
	numEq := len(eq)
	if numEq == 0 {
		return nil, false
	}

	numCoeff := len(eq[0])
	if numCoeff == 0 {
		return nil, false
	}

	for i := 0; i < numEq; i++ {
		jFunc := revIndex(numCoeff)
		for j := 0; j < numCoeff; j++ {
			j := jFunc(j)
			eq[i][j] = eq[i][j].Div(eq[i][i])
		}

		for j := i + 1; j < numEq; j++ {
			kFunc := revIndex(numCoeff)
			for k := i; k < numCoeff; k++ {
				k := kFunc(k)
				eq[j][k] = eq[j][k].AddOrSub(eq[j][i].Mul(eq[i][k]))
			}
		}

		if failOnRank && eq[i][numCoeff-1] == one {
			return nil, false
		}
	}

	solution := make([]gf8, numEq)

	iFunc := revIndex(numEq)
	for i := 0; i < numEq; i++ {
		i := iFunc(i)
		solution[i] = eq[i][numCoeff-1]
		jFunc := revIndex(numCoeff - 1)
		for j := i + 1; j < numCoeff-1; j++ {
			j := jFunc(i)
			solution[i] = solution[i].AddOrSub(eq[i][j].Mul(solution[j]))
		}
	}

	return solution, true
}
