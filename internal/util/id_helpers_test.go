package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLikelyThreadConversationID(t *testing.T) {
	type testCase struct {
		name     string
		id       string
		expected bool
	}
	tests := []testCase{
		{
			name:     "empty string",
			id:       "",
			expected: false,
		},
		{
			name:     "invalid format",
			id:       "not-a-thread-conversation-id",
			expected: false,
		},
		{
			name:     "invalid prefix",
			id:       "20:xxx@thread.tacv2",
			expected: false,
		},
		{
			name:     "missing thread suffix",
			id:       "19:xxx@notthread",
			expected: false,
		},
		{
			name:     "valid thread conversation ID",
			id:       "19:xxx@thread.tacv2",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLikelyThreadConversationID(tt.id)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsLikelyChatID(t *testing.T) {
	type testCase struct {
		name     string
		id       string
		expected bool
	}
	tests := []testCase{
		{
			name:     "empty string",
			id:       "",
			expected: false,
		},
		{
			name:     "invalid format",
			id:       "not-a-chat-id",
			expected: false,
		},
		{
			name:     "invalid prefix",
			id:       "20:xxx@unq.gbl.spaces",
			expected: false,
		},
		{
			name:     "missing unq suffix",
			id:       "19:xxx@notunq",
			expected: false,
		},
		{
			name:     "valid chat ID",
			id:       "19:xxx@unq.gbl.spaces",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLikelyChatID(tt.id)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsLikelyGUID(t *testing.T) {
	type testCase struct {
		name     string
		id       string
		expected bool
	}
	tests := []testCase{
		{
			name:     "empty string",
			id:       "",
			expected: false,
		},
		{
			name:     "invalid GUID format",
			id:       "not-a-guid",
			expected: false,
		},
		{
			name:     "valid GUID",
			id:       "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLikelyGUID(tt.id)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsLikelyEmail_Positive(t *testing.T) {
	s := "user@example.com"
	if !IsLikelyEmail(s) {
		t.Fatalf("expected IsLikelyEmail(%q)=true, got false", s)
	}
}

func TestIsLikelyEmail_Negative(t *testing.T) {
	for _, s := range []string{
		"", "not-an-email", "user@.com", "user@com",
	} {
		if IsLikelyEmail(s) {
			t.Fatalf("expected IsLikelyEmail(%q)=false, got true", s)
		}
	}
}
