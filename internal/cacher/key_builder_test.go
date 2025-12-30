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

func TestNewOneOnOneChatKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewOneOnOneChatKey("user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$direct-chat$:" + hashedEmail

	if got != want {
		t.Fatalf("OneOnOneChatKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestNewGroupChatKey(t *testing.T) {
	got := NewGroupChatKey("Project Alpha")
	want := "$group-chat$:Project Alpha"

	if got != want {
		t.Fatalf("GroupChatKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestNewGroupChatMemberKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewGroupChatMemberKey("chat-123", "user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$group-chat-member$:chat-123:" + hashedEmail

	if got != want {
		t.Fatalf("GroupChatMemberKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestNewTeamMemberKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewChannelMemberKey("team-123", "chan-456", "user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$channel-member$:team-123:chan-456:" + hashedEmail

	if got != want {
		t.Fatalf("MemberKeyBuilder.ToString() = %q, want %q", got, want)
	}
}
