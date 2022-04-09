package ar8t

import "golang.org/x/exp/constraints"

func revIndex[T number](len T) func(i T) T {
	return func(i T) T {
		return len - 1 - i
	}
}

// func revIndex[T number](len, i T) T {
// 	return len - 1 - i
// }

// func revIndexIncl[T number](len, i T) T {
// 	return len - i
// }

// for the time being, to keep up with the ported software

type number interface {
	constraints.Integer | constraints.Float
}
