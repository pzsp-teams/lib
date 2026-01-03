package sender

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/stretchr/testify/require"
)

func TestStatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		want   int
		wantOK bool
	}{
		{
			name:   "nil error -> (0,false)",
			err:    nil,
			want:   0,
			wantOK: false,
		},
		{
			name:   "plain error without StatusCode -> (0,false)",
			err:    errors.New("boom"),
			want:   0,
			wantOK: false,
		},
		{
			name:   "RequestError implements StatusCoder",
			err:    RequestError{Code: http.StatusBadRequest, Message: "bad"},
			want:   http.StatusBadRequest,
			wantOK: true,
		},
		{
			name: "ErrAccessForbidden implements StatusCoder",
			err: ErrAccessForbidden{
				Code:            http.StatusForbidden,
				OriginalMessage: "no scopes",
				ErrData: ErrData{ResourceRefs: map[resources.Resource][]string{
					resources.Team: {"t1"},
				}},
			},
			want:   http.StatusForbidden,
			wantOK: true,
		},
		{
			name: "ErrResourceNotFound implements StatusCoder",
			err: ErrResourceNotFound{
				Code:            http.StatusNotFound,
				OriginalMessage: "missing",
				ErrData: ErrData{ResourceRefs: map[resources.Resource][]string{
					resources.Channel: {"general"},
				}},
			},
			want:   http.StatusNotFound,
			wantOK: true,
		},
		{
			name:   "wrapped with fmt.Errorf(%w) still works",
			err:    fmt.Errorf("outer: %w", RequestError{Code: http.StatusTeapot, Message: "teapot"}),
			want:   http.StatusTeapot,
			wantOK: true,
		},
		{
			name:   "wrapped via Wrap (OpError) still works",
			err:    Wrap("op", RequestError{Code: http.StatusInternalServerError, Message: "oops"}, NewParam(resources.Key("k"), "v")),
			want:   http.StatusInternalServerError,
			wantOK: true,
		},
		{
			name:   "pointer to OpError where underlying is StatusCoder -> works",
			err:    &OpError{Operation: "x", Err: ErrAccessForbidden{Code: http.StatusUnauthorized, OriginalMessage: "nope"}},
			want:   http.StatusUnauthorized,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := StatusCode(tt.err)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
