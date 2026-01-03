package pepper

import (
	"io"
	"os"
	"testing"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func withStdin(t *testing.T, input string) func() {
	t.Helper()

	old := os.Stdin

	r, w, err := os.Pipe()
	require.NoError(t, err)

	_, err = io.WriteString(w, input)
	require.NoError(t, err)

	require.NoError(t, w.Close())
	os.Stdin = r

	return func() {
		os.Stdin = old
		_ = r.Close()
	}
}

func requireKeyringValue(t *testing.T, want string, wantExists bool) {
	t.Helper()

	got, err := keyring.Get(serviceName, userName)
	if !wantExists {
		require.Error(t, err)
		return
	}

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetOrAskPepper(t *testing.T) {
	tests := []struct {
		name         string
		preseed      *string
		stdin        string  
		want         string
		wantErrSub   string 
		wantStored   string
		wantStoredOK bool
	}{
		{
			name:         "keyring has non-empty value -> returns it, no prompt",
			preseed:      util.Ptr("existing-pepper"),
			stdin:        "",
			want:         "existing-pepper",
			wantStored:   "existing-pepper",
			wantStoredOK: true,
		},
		{
			name:         "missing in keyring -> asks stdin, stores and returns trimmed value",
			preseed:      nil,
			stdin:        "  my-pepper \n",
			want:         "my-pepper",
			wantStored:   "my-pepper",
			wantStoredOK: true,
		},
		{
			name:         "keyring has only whitespace -> asks stdin, stores new value",
			preseed:      util.Ptr("   "),
			stdin:        "pep\n",
			want:         "pep",
			wantStored:   "pep",
			wantStoredOK: true,
		},
		{
			name:         "stdin empty -> error 'pepper cannot be empty' and does not store",
			preseed:      nil,
			stdin:        "   \n",
			want:         "",
			wantErrSub:   "pepper cannot be empty",
			wantStoredOK: false,
		},
		{
			name:         "stdin read error (EOF before newline) -> returns 'reading pepper' error and does not store",
			preseed:      nil,
			stdin:        "no-newline",
			want:         "",
			wantErrSub:   "reading pepper",
			wantStoredOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring.MockInit()

			if tt.preseed != nil {
				require.NoError(t, keyring.Set(serviceName, userName, *tt.preseed))
			}

			var restore func()
			if tt.stdin != "" {
				restore = withStdin(t, tt.stdin)
				defer restore()
			}

			got, err := GetOrAskPepper()
			if tt.wantErrSub != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrSub)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)

			requireKeyringValue(t, tt.wantStored, tt.wantStoredOK)
		})
	}
}

