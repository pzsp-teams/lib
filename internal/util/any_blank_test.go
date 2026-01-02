package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnyBlank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want bool
	}{
		{
			name: "no args -> false",
			in:   nil,
			want: false,
		},
		{
			name: "single non-blank -> false",
			in:   []string{"abc"},
			want: false,
		},
		{
			name: "single blank -> true",
			in:   []string{""},
			want: true,
		},
		{
			name: "single whitespace-only -> true",
			in:   []string{"   \t\n"},
			want: true,
		},
		{
			name: "multiple non-blank -> false",
			in:   []string{"a", "b", "c"},
			want: false,
		},
		{
			name: "contains empty among non-blank -> true",
			in:   []string{"a", "", "c"},
			want: true,
		},
		{
			name: "contains whitespace-only among non-blank -> true",
			in:   []string{"a", "   ", "c"},
			want: true,
		},
		{
			name: "all blank -> true",
			in:   []string{"", " ", "\t"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := AnyBlank(tt.in...)
			require.Equal(t, tt.want, got)
		})
	}
}
