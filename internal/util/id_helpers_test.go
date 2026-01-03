package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsLikelyThreadConversationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   \n\t  ", false},
		{"invalid format", "not-a-thread-conversation-id", false},
		{"invalid prefix", "20:xxx@thread.tacv2", false},
		{"missing @thread. part", "19:xxx@notthread", false},
		{"valid", "19:xxx@thread.tacv2", true},
		{"valid with leading/trailing spaces", "  19:xxx@thread.tacv2 \n", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, IsLikelyThreadConversationID(tt.in))
		})
	}
}

func TestIsLikelyChatID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   ", false},
		{"invalid format", "not-a-chat-id", false},
		{"invalid prefix", "20:xxx@unq.gbl.spaces", false},
		{"missing @unq. part", "19:xxx@notunq", false},
		{"valid", "19:xxx@unq.gbl.spaces", true},
		{"valid with leading/trailing spaces", "\t19:xxx@unq.gbl.spaces  ", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, IsLikelyChatID(tt.in))
		})
	}
}

func TestIsLikelyGUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   ", false},
		{"invalid format", "not-a-guid", false},
		{"valid lowercase", "123e4567-e89b-12d3-a456-426614174000", true},
		{"valid uppercase", "123E4567-E89B-12D3-A456-426614174000", true},
		{"too short", "123e4567-e89b-12d3-a456-42661417400", false},
		{"missing dashes", "123e4567e89b12d3a456426614174000", false},
		{"valid but with spaces -> true", " 123e4567-e89b-12d3-a456-426614174000 ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, IsLikelyGUID(tt.in))
		})
	}
}

func TestIsLikelyEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   ", false},
		{"missing at", "not-an-email", false},
		{"missing domain label", "user@.com", false},
		{"missing dot tld", "user@com", false},
		{"valid simple", "user@example.com", true},
		{"valid plus tag", "user+tag@example.com", true},
		{"valid dots", "first.last@example.com", true},
		{"valid subdomain", "user@mail.example.com", true},
		{"valid with leading/trailing spaces (trimmed)", "  user@example.com \n", true},
		{"invalid double at", "user@@example.com", false},
		{"invalid tld too short", "user@example.c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, IsLikelyEmail(tt.in))
		})
	}
}
