package pepper

import (
	"errors"
	"testing"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestGetPepper(t *testing.T) {
	tests := []struct {
		name      string
		preseed   *string
		want      string
		wantErrIs error
	}{
		{
			name:      "missing in keyring -> ErrPepperNotSet",
			preseed:   nil,
			want:      "",
			wantErrIs: ErrPepperNotSet,
		},
		{
			name:      "present but empty -> ErrPepperNotSet",
			preseed:   util.Ptr(""),
			want:      "",
			wantErrIs: ErrPepperNotSet,
		},
		{
			name:      "present but whitespace -> ErrPepperNotSet",
			preseed:   util.Ptr("   \t\n"),
			want:      "",
			wantErrIs: ErrPepperNotSet,
		},
		{
			name:      "present -> returns trimmed value",
			preseed:   util.Ptr("  my-pepper \n"),
			want:      "my-pepper",
			wantErrIs: nil,
		},
		{
			name:      "present already trimmed -> returns as-is",
			preseed:   util.Ptr("pep"),
			want:      "pep",
			wantErrIs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring.MockInit()

			if tt.preseed != nil {
				require.NoError(t, keyring.Set(serviceName, userName, *tt.preseed))
			}

			got, err := GetPepper()
			if tt.wantErrIs != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErrIs), "expected errors.Is(err, %v) to be true, got err=%v", tt.wantErrIs, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSetPepper(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantStored    string
		wantErrSubstr string
		wantExists    bool
	}{
		{
			name:       "valid -> stores trimmed value",
			input:      "  abc \n",
			wantStored: "abc",
			wantExists: true,
		},
		{
			name:          "empty -> error and does not store",
			input:         "",
			wantErrSubstr: "pepper cannot be empty",
			wantExists:    false,
		},
		{
			name:          "whitespace -> error and does not store",
			input:         "   \t\n",
			wantErrSubstr: "pepper cannot be empty",
			wantExists:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring.MockInit()

			err := SetPepper(tt.input)
			if tt.wantErrSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrSubstr)

				_, getErr := keyring.Get(serviceName, userName)
				require.Error(t, getErr)
				require.True(t, errors.Is(getErr, keyring.ErrNotFound))
				return
			}

			require.NoError(t, err)

			got, getErr := keyring.Get(serviceName, userName)
			require.NoError(t, getErr)
			require.Equal(t, tt.wantStored, got)
		})
	}
}

func TestPepperExists(t *testing.T) {
	tests := []struct {
		name    string
		preseed *string
		want    bool
	}{
		{
			name:    "missing -> false",
			preseed: nil,
			want:    false,
		},
		{
			name:    "present but empty -> false",
			preseed: util.Ptr(""),
			want:    false,
		},
		{
			name:    "present but whitespace -> false",
			preseed: util.Ptr("   "),
			want:    false,
		},
		{
			name:    "present -> true",
			preseed: util.Ptr("pep"),
			want:    true,
		},
		{
			name:    "present with surrounding whitespace -> true",
			preseed: util.Ptr("  pep \n"),
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring.MockInit()

			if tt.preseed != nil {
				require.NoError(t, keyring.Set(serviceName, userName, *tt.preseed))
			}

			got, err := PepperExists()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSetGetExists_Integration(t *testing.T) {
	keyring.MockInit()

	require.NoError(t, SetPepper("  xyz  "))

	ok, err := PepperExists()
	require.NoError(t, err)
	require.True(t, ok)

	got, err := GetPepper()
	require.NoError(t, err)
	require.Equal(t, "xyz", got)
}
