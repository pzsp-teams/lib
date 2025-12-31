package cacher

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewTeamKey(t *testing.T) {
	got := NewTeamKey("my-team")
	want := "$team$:my-team"
	assert.Equal(t, want, got, "NewTeamKey() should return the correct key string")
}

func TestNewChannelKey(t *testing.T) {
	got := NewChannelKey("team-123", "general")
	want := "$channel$:team-123:general"
	assert.Equal(t, want, got, "NewChannelKey() should return the correct key string")
}

func TestNewOneOnOneChatKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewOneOnOneChatKey("user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$direct-chat$:" + hashedEmail
	assert.Equal(t, want, got, "NewOneOnOneChatKey() should return the correct key string")
}

func TestNewGroupChatKey(t *testing.T) {
	got := NewGroupChatKey("Project Alpha")
	want := "$group-chat$:Project Alpha"
	assert.Equal(t, want, got, "NewGroupChatKey() should return the correct key string")
}

func TestNewGroupChatMemberKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewGroupChatMemberKey("chat-123", "user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$group-chat-member$:chat-123:" + hashedEmail
	assert.Equal(t, want, got, "NewGroupChatMemberKey() should return the correct key string")
}

func TestNewChannelMemberKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewChannelMemberKey("team-123", "chan-456", "user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$channel-member$:team-123:chan-456:" + hashedEmail
	assert.Equal(t, want, got, "NewChannelMemberKey() should return the correct key string")
}

func TestNewTeamMemberKey(t *testing.T) {
	testPepper := "test-pepper"
	got := NewTeamMemberKey("team-123", "user@example.com", &testPepper)
	hashedEmail := util.HashWithPepper(testPepper, "user@example.com")
	want := "$team-member$:team-123:" + hashedEmail

	if got != want {
		t.Fatalf("TeamMemberKeyBuilder.ToString() = %q, want %q", got, want)
	}
}
