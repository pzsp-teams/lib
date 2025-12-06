package util

import (
	"testing"
)

func TestDeref_NilReturnsEmpty(t *testing.T) {
	var text *string = nil
	if got := Deref(text); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestDeref_NonNil(t *testing.T) {
	s := "hello"
	if got := Deref(&s); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}
