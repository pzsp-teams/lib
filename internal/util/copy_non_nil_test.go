package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyNonNil_Ints(t *testing.T) {
	a, b, c := 1, 2, 3

	tests := []struct {
		name  string
		input []*int
		want  []int
	}{
		{
			name:  "nil slice",
			input: nil,
			want:  []int{},
		},
		{
			name:  "empty slice",
			input: []*int{},
			want:  []int{},
		},
		{
			name:  "all non-nil",
			input: []*int{&a, &b, &c},
			want:  []int{1, 2, 3},
		},
		{
			name:  "mixed with nils",
			input: []*int{nil, &a, nil, &c},
			want:  []int{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CopyNonNil(tt.input)

			assert.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i], got[i], "mismatch at index %d", i)
			}
		})
	}
}

func TestCopyNonNil_NoAliasing(t *testing.T) {
	x := 10
	input := []*int{&x}

	out := CopyNonNil(input)

	assert.Equal(t, len(out), 1)
	assert.Equal(t, out[0], 10)

	out[0] = 42

	assert.Equal(t, 10, x, "input value should not be affected by changes to output")
}
