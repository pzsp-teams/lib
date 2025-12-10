package util

import (
	"strconv"
	"testing"
)

func TestMapSlices_Basic(t *testing.T) {
	in := []int{1, 2, 3}
	out := MapSlices(in, func(v int) string {
		return "v=" + strconv.Itoa(v)
	})

	want := []string{"v=1", "v=2", "v=3"}
	if len(out) != len(want) {
		t.Fatalf("expected len=%d, got %d (out=%v)", len(want), len(out), out)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Errorf("at %d expected %q, got %q", i, want[i], out[i])
		}
	}
}

func TestMapSlices_EmptyAndNil(t *testing.T) {
	var nilSlice []int
	emptySlice := []int{}

	outNil := MapSlices(nilSlice, strconv.Itoa)
	outEmpty := MapSlices(emptySlice, strconv.Itoa)

	if len(outNil) != 0 {
		t.Errorf("expected len=0 for nil slice, got %d", len(outNil))
	}
	if len(outEmpty) != 0 {
		t.Errorf("expected len=0 for empty slice, got %d", len(outEmpty))
	}
}

func TestMapSlices_MapperCalledForEachElement(t *testing.T) {
	in := []int{10, 20, 30}
	var calls int

	out := MapSlices(in, func(v int) int {
		calls++
		return v * 2
	})

	if calls != len(in) {
		t.Errorf("expected mapper to be called %d times, got %d", len(in), calls)
	}
	if out[0] != 20 || out[1] != 40 || out[2] != 60 {
		t.Errorf("unexpected output: %v", out)
	}
}
