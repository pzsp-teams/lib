package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSlices(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		fn   func(int) string
		want []string
	}{
		{
			name: "basic",
			in:   []int{1, 2, 3},
			fn: func(v int) string {
				return "v=" + strconv.Itoa(v)
			},
			want: []string{"v=1", "v=2", "v=3"},
		},
		{
			name: "nil slice",
			in:   nil,
			fn:   strconv.Itoa,
			want: []string{},
		},
		{
			name: "empty slice",
			in:   []int{},
			fn:   strconv.Itoa,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := MapSlices(tt.in, tt.fn)

			require.Equal(t, len(tt.want), len(out))

			for i := range tt.want {
				assert.Equal(t, tt.want[i], out[i])
			}
		})
	}
}

func TestMapSlices_MapperCalledForEachElement(t *testing.T) {
	in := []int{10, 20, 30}
	var calls int

	out := MapSlices(in, func(v int) int {
		calls++
		return v * 2
	})

	assert.Equal(t, len(in), calls, "mapper should be called for each element")
	for i, elem := range in {
		expected := elem * 2
		assert.Equal(t, expected, out[i], "different output value at index %d", i)
	}
}
