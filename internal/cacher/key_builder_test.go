package cacher

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/require"
)

func TestKeyBuilders_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{
			name: "NewTeamKey formats key",
			got: func() string {
				return NewTeamKey("my-team")
			},
			want: "$team$:my-team",
		},
		{
			name: "NewTeamKey trims whitespace",
			got: func() string {
				return NewTeamKey("  my-team  ")
			},
			want: "$team$:my-team",
		},
		{
			name: "NewChannelKey formats key",
			got: func() string {
				return NewChannelKey("team-123", "general")
			},
			want: "$channel$:team-123:general",
		},
		{
			name: "NewChannelKey trims whitespace parts",
			got: func() string {
				return NewChannelKey(" team-123 ", "  general  ")
			},
			want: "$channel$:team-123:general",
		},
		{
			name: "NewGroupChatKey formats key",
			got: func() string {
				return NewGroupChatKey("Project Alpha")
			},
			want: "$group-chat$:Project Alpha",
		},
		{
			name: "NewGroupChatKey trims whitespace",
			got: func() string {
				return NewGroupChatKey("  Project Alpha  ")
			},
			want: "$group-chat$:Project Alpha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.got())
		})
	}
}

func TestKeyBuilders_HashedRefs_WithExplicitPepper(t *testing.T) {
	t.Parallel()

	pep := "test-pepper"

	tests := []struct {
		name string
		got  func() string
		want func() string
	}{
		{
			name: "NewOneOnOneChatKey hashes user ref",
			got: func() string {
				return NewOneOnOneChatKey("user@example.com", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$direct-chat$:" + hashed
			},
		},
		{
			name: "NewOneOnOneChatKey trims user ref before hashing",
			got: func() string {
				return NewOneOnOneChatKey("  user@example.com  ", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$direct-chat$:" + hashed
			},
		},
		{
			name: "NewGroupChatMemberKey hashes user ref and includes chat id",
			got: func() string {
				return NewGroupChatMemberKey("chat-123", "user@example.com", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$group-chat-member$:chat-123:" + hashed
			},
		},
		{
			name: "NewChannelMemberKey hashes user ref and includes team + channel",
			got: func() string {
				return NewChannelMemberKey("team-123", "chan-456", "user@example.com", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$channel-member$:team-123:chan-456:" + hashed
			},
		},
		{
			name: "NewTeamMemberKey hashes user ref and includes team id",
			got: func() string {
				return NewTeamMemberKey("team-123", "user@example.com", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$team-member$:team-123:" + hashed
			},
		},
		{
			name: "NewTeamMemberKey trims user ref before hashing",
			got: func() string {
				return NewTeamMemberKey("team-123", "  user@example.com  ", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$team-member$:team-123:" + hashed
			},
		},
		{
			name: "NewChannelMemberKey trims team/channel parts too",
			got: func() string {
				return NewChannelMemberKey(" team-123 ", " chan-456 ", " user@example.com ", &pep)
			},
			want: func() string {
				hashed := util.HashWithPepper(pep, "user@example.com")
				return "$channel-member$:team-123:chan-456:" + hashed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want(), tt.got())
		})
	}
}
