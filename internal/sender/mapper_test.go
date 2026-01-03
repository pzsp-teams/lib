package sender

import (
	"errors"
	"net/http"
	"testing"

	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/stretchr/testify/require"
)

func TestWithResource_And_WithResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []Option
		want map[resources.Resource][]string
	}{
		{
			name: "WithResource initializes map and appends single ref",
			opts: []Option{WithResource(resources.Team, "t1")},
			want: map[resources.Resource][]string{
				resources.Team: {"t1"},
			},
		},
		{
			name: "WithResources initializes map and appends many refs",
			opts: []Option{WithResources(resources.Channel, []string{"c1", "c2"})},
			want: map[resources.Resource][]string{
				resources.Channel: {"c1", "c2"},
			},
		},
		{
			name: "mixing options appends and preserves existing",
			opts: []Option{
				WithResource(resources.Team, "t1"),
				WithResources(resources.Team, []string{"t2", "t3"}),
				WithResource(resources.Channel, "general"),
				WithResource(resources.Team, "t4"),
			},
			want: map[resources.Resource][]string{
				resources.Team:    {"t1", "t2", "t3", "t4"},
				resources.Channel: {"general"},
			},
		},
		{
		name: "WithResources with nil slice creates key with empty value",
		opts: []Option{WithResources(resources.Team, nil)},
		want: map[resources.Resource][]string{
			resources.Team: nil, // albo []string{}
		},
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d ErrData
			for _, opt := range tt.opts {
				opt(&d)
			}

			if len(tt.want) == 0 {
				if d.ResourceRefs == nil {
					return
				}
				require.Len(t, d.ResourceRefs, 0)
				return
			}

			require.NotNil(t, d.ResourceRefs)
			require.Equal(t, tt.want, d.ResourceRefs)
		})
	}
}

func TestMapError(t *testing.T) {
	t.Parallel()

	t.Run("nil request error -> nil", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, MapError(nil))
	})

	t.Run("StatusForbidden -> ErrAccessForbidden with data", func(t *testing.T) {
		t.Parallel()

		in := &RequestError{Code: http.StatusForbidden, Message: "no scopes"}
		err := MapError(in,
			WithResource(resources.Team, "t1"),
			WithResources(resources.User, []string{"u1", "u2"}),
		)

		var af *ErrAccessForbidden
		require.ErrorAs(t, err, &af)

		require.Equal(t, http.StatusForbidden, af.Code)
		require.Equal(t, "no scopes", af.OriginalMessage)
		require.Equal(t, map[resources.Resource][]string{
			resources.Team: {"t1"},
			resources.User: {"u1", "u2"},
		}, af.ResourceRefs)

		require.True(t, errors.Is(err, ErrAccessForbidden{Code: http.StatusForbidden}))
		require.False(t, errors.Is(err, ErrAccessForbidden{Code: http.StatusNotFound}))
	})

	t.Run("StatusNotFound -> ErrResourceNotFound with data", func(t *testing.T) {
		t.Parallel()

		in := &RequestError{Code: http.StatusNotFound, Message: "missing"}
		err := MapError(in, WithResource(resources.Channel, "general"))

		var nf *ErrResourceNotFound
		require.ErrorAs(t, err, &nf)

		require.Equal(t, http.StatusNotFound, nf.Code)
		require.Equal(t, "missing", nf.OriginalMessage)
		require.Equal(t, map[resources.Resource][]string{
			resources.Channel: {"general"},
		}, nf.ResourceRefs)

		require.True(t, errors.Is(err, ErrResourceNotFound{Code: http.StatusNotFound}))
		require.False(t, errors.Is(err, ErrResourceNotFound{Code: http.StatusBadRequest}))
	})

	t.Run("default -> returns copy of RequestError (same fields, different pointer)", func(t *testing.T) {
		t.Parallel()

		in := &RequestError{Code: http.StatusBadRequest, Message: "bad"}
		err := MapError(in, WithResource(resources.Team, "t1")) 

		var out *RequestError
		require.ErrorAs(t, err, &out)

		require.Equal(t, in.Code, out.Code)
		require.Equal(t, in.Message, out.Message)
		require.NotSame(t, in, out, "MapError should return a copy, not the same pointer")
	})
}
