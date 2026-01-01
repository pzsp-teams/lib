package util

import "testing"

func TestMemberRole(t *testing.T) {
	if got := MemberRole(true); got[0] != "owner" {
		t.Fatalf("expected owner, got %q", got)
	}
	if got := MemberRole(false); len(got) != 0 {
		t.Fatalf("expected empty slice, got %q", got)
	}
}
