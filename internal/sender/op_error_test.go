package sender

import (
	"errors"
	"testing"

	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/stretchr/testify/require"
)

func TestNewParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   resources.Key
		value []string
		want  Param
	}{
		{
			name:  "no values -> empty slice",
			key:   resources.Key("k"),
			value: nil,
			want:  Param{Key: resources.Key("k"), Value: nil},
		},
		{
			name:  "single value",
			key:   resources.Key("k"),
			value: []string{"v1"},
			want:  Param{Key: resources.Key("k"), Value: []string{"v1"}},
		},
		{
			name:  "multiple values",
			key:   resources.Key("k"),
			value: []string{"v1", "v2"},
			want:  Param{Key: resources.Key("k"), Value: []string{"v1", "v2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewParam(tt.key, tt.value...)
			require.Equal(t, tt.want.Key, got.Key)
			require.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	t.Run("nil err -> nil", func(t *testing.T) {
		t.Parallel()

		got := Wrap("op", nil, NewParam(resources.Key("k"), "v"))
		require.Nil(t, got)
	})

	t.Run("non-nil err -> returns *OpError", func(t *testing.T) {
		t.Parallel()

		base := errors.New("boom")
		got := Wrap("op", base, NewParam(resources.Key("k"), "v1", "v2"))

		var oe *OpError
		require.ErrorAs(t, got, &oe)

		require.Equal(t, "op", oe.Operation)
		require.Len(t, oe.Params, 1)
		require.Equal(t, resources.Key("k"), oe.Params[0].Key)
		require.Equal(t, []string{"v1", "v2"}, oe.Params[0].Value)
		require.ErrorIs(t, got, base)
	})
}

func TestOpError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *OpError
		want string
	}{
		{
			name: "no params",
			in: &OpError{
				Operation: "ListTeams",
				Params:    nil,
				Err:       errors.New("boom"),
			},
			want: "Error in ListTeams: boom",
		},
		{
			name: "with params",
			in: &OpError{
				Operation: "GetChannel",
				Params: []Param{
					NewParam(resources.Key("teamID"), "t1"),
					NewParam(resources.Key("channelRef"), "general"),
				},
				Err: errors.New("not found"),
			},
			want: "Error in GetChannel [teamID=[t1], channelRef=[general]]: not found",
		},
		{
			name: "param with multiple values",
			in: &OpError{
				Operation: "ResolveUser",
				Params: []Param{
					NewParam(resources.Key("userRef"), "a@x.com", "b@x.com"),
				},
				Err: errors.New("ambiguous"),
			},
			want: "Error in ResolveUser [userRef=[a@x.com b@x.com]]: ambiguous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.in.Error())
		})
	}
}

func TestOpError_Unwrap(t *testing.T) {
	t.Parallel()

	base := errors.New("root")
	oe := &OpError{
		Operation: "Op",
		Err:       base,
	}

	require.Equal(t, base, oe.Unwrap())
	require.ErrorIs(t, oe, base)
}
