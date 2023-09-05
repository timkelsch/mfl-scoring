package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundFloat(t *testing.T) {
	type args struct {
		value     float64
		precision uint
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "small", args: args{-0.34567, 3}, want: -0.346},
		{name: "large", args: args{4923487768956.98234779857, 8}, want: 4923487768956.98234780},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundFloat(tt.args.value, tt.args.precision)
			assert.Equal(t, tt.want, got)
		})
	}
}
