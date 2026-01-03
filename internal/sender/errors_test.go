package sender

import (
	"errors"
	"net/http"
	"testing"

	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/stretchr/testify/require"
)

func requireContainsAll(t *testing.T, got string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		require.Contains(t, got, p)
	}
}

func TestRequestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   RequestError
		want string
	}{
		{
			name: "formats with code and message",
			in:   RequestError{Code: http.StatusForbidden, Message: "Forbidden"},
			want: "[CODE: 403]: Forbidden",
		},
		{
			name: "empty message still formats",
			in:   RequestError{Code: http.StatusBadRequest, Message: ""},
			want: "[CODE: 400]: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.in.Error())
		})
	}
}

func TestErrData_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    *ErrData
		want  string
		parts []string
	}{
		{
			name: "nil map -> empty string",
			in:   &ErrData{ResourceRefs: nil},
			want: "",
		},
		{
			name: "empty map -> empty string",
			in:   &ErrData{ResourceRefs: map[resources.Resource][]string{}},
			want: "",
		},
		{
			name: "single entry one ref -> deterministic",
			in: &ErrData{ResourceRefs: map[resources.Resource][]string{
				resources.Team: {"z1"},
			}},
			want: "TEAM(z1)",
		},
		{
			name: "single entry many refs -> deterministic join",
			in: &ErrData{ResourceRefs: map[resources.Resource][]string{
				resources.Team: {"z1", "z2"},
			}},
			want: "TEAM(z1,z2)",
		},
		{
			name: "multiple entries -> contains all segments (order not guaranteed)",
			in: &ErrData{ResourceRefs: map[resources.Resource][]string{
				resources.Team:    {"z1"},
				resources.Channel: {"general"},
			}},
			parts: []string{"TEAM(z1)", "CHANNEL(general)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.in.String()

			if tt.want != "" || (tt.in != nil && tt.in.ResourceRefs != nil && len(tt.in.ResourceRefs) == 0) {
				require.Equal(t, tt.want, got)
				return
			}
			requireContainsAll(t, got, tt.parts...)
		})
	}
}

func TestErrAccessForbidden_Error_And_Is(t *testing.T) {
	t.Parallel()

	base := ErrAccessForbidden{
		Code:            http.StatusForbidden,
		OriginalMessage: "Bad scopes",
		ErrData: ErrData{ResourceRefs: map[resources.Resource][]string{
			resources.Team:    {"z1"},
			resources.Channel: {"general"},
		}},
	}

	t.Run("Error contains stable parts", func(t *testing.T) {
		t.Parallel()

		got := base.Error()
		requireContainsAll(t, got,
			"[CODE: 403]:",
			"access forbidden to one or more resources among:",
			"(Bad scopes)",
			"TEAM(z1)",
			"CHANNEL(general)",
		)
	})

	t.Run("errors.Is behavior (value vs pointer target)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			err    error
			target error
			want   bool
		}{
			{
				name:   "matches same type+code (value err, value target)",
				err:    base,
				target: ErrAccessForbidden{Code: http.StatusForbidden},
				want:   true,
			},
			{
				name:   "does not match different code",
				err:    base,
				target: ErrAccessForbidden{Code: http.StatusNotFound},
				want:   false,
			},
			{
				name:   "does not match different type even if same code",
				err:    base,
				target: ErrResourceNotFound{Code: http.StatusForbidden},
				want:   false,
			},
			{
				name:   "target as pointer does NOT match (Is asserts value type)",
				err:    base,
				target: &ErrAccessForbidden{Code: http.StatusForbidden},
				want:   false,
			},
			{
				name:   "err as pointer still matches value target (receiver is value)",
				err:    &base,
				target: ErrAccessForbidden{Code: http.StatusForbidden},
				want:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				require.Equal(t, tt.want, errors.Is(tt.err, tt.target))
			})
		}
	})
}

func TestErrResourceNotFound_Error_And_Is(t *testing.T) {
	t.Parallel()

	base := ErrResourceNotFound{
		Code:            http.StatusNotFound,
		OriginalMessage: "Not found",
		ErrData: ErrData{ResourceRefs: map[resources.Resource][]string{
			resources.Channel: {"general"},
		}},
	}

	t.Run("Error formats deterministically for single entry", func(t *testing.T) {
		t.Parallel()

		want := "[CODE: 404]: one or more resources not found among: CHANNEL(general) (Not found)"
		require.Equal(t, want, base.Error())
	})

	t.Run("errors.Is behavior (value vs pointer target)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			err    error
			target error
			want   bool
		}{
			{
				name:   "matches same type+code",
				err:    base,
				target: ErrResourceNotFound{Code: http.StatusNotFound},
				want:   true,
			},
			{
				name:   "does not match different code",
				err:    base,
				target: ErrResourceNotFound{Code: http.StatusBadRequest},
				want:   false,
			},
			{
				name:   "does not match different type",
				err:    base,
				target: ErrAccessForbidden{Code: http.StatusNotFound},
				want:   false,
			},
			{
				name:   "target as pointer does NOT match (Is asserts value type)",
				err:    base,
				target: &ErrResourceNotFound{Code: http.StatusNotFound},
				want:   false,
			},
			{
				name:   "err as pointer still matches value target",
				err:    &base,
				target: ErrResourceNotFound{Code: http.StatusNotFound},
				want:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				require.Equal(t, tt.want, errors.Is(tt.err, tt.target))
			})
		}
	})
}
