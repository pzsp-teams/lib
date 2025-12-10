package util

import "testing"

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
		tt := tt 
		t.Run(tt.name, func(t *testing.T) {
			got := CopyNonNil(tt.input)

			if len(got) != len(tt.want) {
				t.Fatalf("expected len=%d, got %d (got=%v, want=%v)", len(tt.want), len(got), got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("at index %d expected %d, got %d", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestCopyNonNil_NoAliasing(t *testing.T) {
	x := 10
	input := []*int{&x}

	out := CopyNonNil(input)

	if len(out) != 1 || out[0] != 10 {
		t.Fatalf("unexpected result: %#v", out)
	}

	out[0] = 42

	if x != 10 {
		t.Fatalf("expected original x to remain 10, got %d", x)
	}
}
