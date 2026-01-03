package util

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashWithPepper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pepper string
		value  string
		want   string
	}{
		{
			name:   "known vector",
			pepper: "pepper123",
			value:  "user@example.com",
			want:   "3cadad0f8d29a1acd80eb42d09b809c554fb7e4bb70051e67193369d56abc021",
		},
		{
			name:   "empty pepper and value still produces valid sha256 hex",
			pepper: "",
			value:  "",
			want:   "",
		},
		{
			name:   "unicode input still produces valid sha256 hex",
			pepper: "pieprz🔒",
			value:  "użytkownik@example.com",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := HashWithPepper(tt.pepper, tt.value)

			require.Len(t, got, 64, "sha256 hex length should be 64")
			_, err := hex.DecodeString(got)
			require.NoError(t, err, "hash should be valid hex")

			if tt.want != "" {
				require.Equal(t, tt.want, got)
			}
		})
	}

	t.Run("changes when pepper changes (same value)", func(t *testing.T) {
		t.Parallel()

		v := "user@example.com"
		h1 := HashWithPepper("pepper1", v)
		h2 := HashWithPepper("pepper2", v)
		require.NotEqual(t, h1, h2)
	})

	t.Run("changes when value changes (same pepper)", func(t *testing.T) {
		t.Parallel()

		p := "pepper123"
		h1 := HashWithPepper(p, "value1")
		h2 := HashWithPepper(p, "value2")
		require.NotEqual(t, h1, h2)
	})

	t.Run("deterministic (same inputs -> same output)", func(t *testing.T) {
		t.Parallel()

		p := "pepper123"
		v := "user@example.com"
		h1 := HashWithPepper(p, v)
		h2 := HashWithPepper(p, v)
		require.Equal(t, h1, h2)
	})
}
