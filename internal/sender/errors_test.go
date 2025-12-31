package sender

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestRequestError(t *testing.T) {
	e := RequestError{Code: http.StatusForbidden, Message: "Forbidden"}
	got := e.Error()
	want := "[CODE: 403]: Forbidden"
	assert.Equal(t, want, got)
}

func TestErrData(t *testing.T) {
	tests := []struct {
		name string
		ed   ErrData
		want []string
	}{
		{
			name: "empty map",
			ed:   ErrData{ResourceRefs: map[resources.Resource]string{}},
			want: []string{},
		},
		{
			name: "single entry",
			ed:   ErrData{ResourceRefs: map[resources.Resource]string{resources.Team: "z1"}},
			want: []string{"TEAM(z1)"},
		},
		{
			name: "multiple entries",
			ed: ErrData{ResourceRefs: map[resources.Resource]string{
				resources.Team:    "z1",
				resources.Channel: "general",
			}},
			want: []string{"TEAM(z1)", "CHANNEL(general)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ed.String()
			if len(tt.want) != 0 {
				for _, w := range tt.want {
					assert.True(t, strings.Contains(got, w))
				}
			}
		})
	}
}

func TestErrAccessForbidden(t *testing.T) {
	e := ErrAccessForbidden{
		Code:            http.StatusForbidden,
		OriginalMessage: "Bad scopes",
		ErrData: ErrData{ResourceRefs: map[resources.Resource]string{
			resources.Team: "z1",
		}},
	}

	got := e.Error()
	want := "[CODE: 403]: access forbidden to one or more resources among: TEAM(z1) (Bad scopes)"
	assert.Equal(t, want, got)

	assert.ErrorIs(t, e, ErrAccessForbidden{Code: http.StatusForbidden})
	assert.NotErrorIs(t, e, ErrAccessForbidden{Code: http.StatusNotFound})
	assert.NotErrorIs(t, e, ErrResourceNotFound{Code: http.StatusForbidden})
}

func TestErrResourceNotFound_Table(t *testing.T) {
	e := ErrResourceNotFound{
		Code:            http.StatusNotFound,
		OriginalMessage: "Bad scopes",
		ErrData: ErrData{ResourceRefs: map[resources.Resource]string{
			resources.Channel: "general",
		}},
	}

	got := e.Error()
	want := "[CODE: 404]: one or more resources not found among: CHANNEL(general) (Bad scopes)"
	assert.Equal(t, want, got)

	assert.ErrorIs(t, e, ErrResourceNotFound{Code: http.StatusNotFound})
	assert.NotErrorIs(t, e, ErrResourceNotFound{Code: http.StatusBadRequest})
	assert.NotErrorIs(t, e, ErrAccessForbidden{Code: http.StatusNotFound})
}
