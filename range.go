package ar8t

import "golang.org/x/exp/constraints"

type _range[T constraints.Integer | constraints.Float] struct {
	Min, Max T
}

func (r _range[T]) Reverse() _range[T] {
	return _range[T]{r.Max, r.Min}
}

func (r _range[T]) Iter() *rangeIter[T] {
	return &rangeIter[T]{r, r.Min}
}

type rangeIter[T constraints.Integer | constraints.Float] struct {
	_range[T]
	current T
}

//[Min, Max)
func (r *rangeIter[T]) Next() bool {
	return r.current < r.Max
}

func (r *rangeIter[T]) Value() T {
	r.current++
	return r.current - 1
}

type Iterator[T any] interface {
	Next() bool
	Value() T
}

type repeatIter[T any] struct {
	Val T
}

func (repeatIter[_]) Next() bool {
	return true
}

func (i repeatIter[T]) Value() T {
	return i.Val
}

type zipIter[T1, T2 any] struct {
	i1 Iterator[T1]
	i2 Iterator[T2]
}

func (zipIter[_, _]) Next() bool {
	return true
}

func (i zipIter[T1, T2]) Value() (T1, T2) {
	return i.i1.Value(), i.i2.Value()
}

func zip[T1, T2 any](i1 Iterator[T1], i2 Iterator[T2]) zipIter[T1, T2] {
	return zipIter[T1, T2]{i1, i2}
}
