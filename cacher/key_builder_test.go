package cacher

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/util"
)

func TestNewTeamKey(t *testing.T) {
	got := NewTeamKey("my-team")
	want := "$team$:my-team"

	if got != want {
		t.Fatalf("TeamKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestNewChannelKey(t *testing.T) {
	got := NewChannelKey("team-123", "general")
	want := "$channel$:team-123:general"

	if got != want {
		t.Fatalf("ChannelKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestNewMemberKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewMemberKey("user@example.com", "team-123", "chan-456", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$member$:team-123:chan-456:" + hashedEmail

	if got != want {
		t.Fatalf("MemberKeyBuilder.ToString() = %q, want %q", got, want)
	}
}
