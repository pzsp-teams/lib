package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemberRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		isOwner bool
		want    []string
	}{
		{
			name:    "owner true returns [owner]",
			isOwner: true,
			want:    []string{"owner"},
		},
		{
			name:    "owner false returns empty slice",
			isOwner: false,
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := MemberRole(tt.isOwner)
			require.Equal(t, tt.want, got)
		})
	}
}
