package ar8t

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_revIndex(t *testing.T) {
	type args struct {
		len int
	}
	tests := []struct {
		name string
		args args
		res  map[int]int //map[index]want
	}{
		{
			name: "from zero to n",
			args: args{10},
			res:  map[int]int{0: 9, 4: 5, 9: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := revIndex(tt.args.len)
			for index, want := range tt.res {
				got := fn(index)
				assert.Equal(t, want, got)
			}
		})
	}
}
