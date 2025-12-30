package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeref(t *testing.T) {
	type testCase struct {
		name string
		got  func() any
		want any
	}

	tests := []testCase{
		{
			name: "nil *int",
			got:  func() any { return Deref((*int)(nil)) },
			want: 0,
		},
		{
			name: "non-nil *int",
			got:  func() any { return Deref(Ptr(42)) },
			want: 42,
		},
		{
			name: "nil *string",
			got:  func() any { return Deref((*string)(nil)) },
			want: "",
		},
		{
			name: "non-nil *string",
			got:  func() any { return Deref(Ptr("hello")) },
			want: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.got()
			assert.Equal(t, tt.want, got)
		})
	}
}
