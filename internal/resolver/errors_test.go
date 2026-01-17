package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	type tc struct {
		name string
		err  error
		want string
	}

	tests := []tc{
		{
			name: "resourcesNotAvailableError",
			err: &resourcesNotAvailableError{
				resourceType: resources.Team,
			},
			want: fmt.Sprintf(
				"cannot resolve %s: resources not available",
				resources.Team,
			),
		},
		{
			name: "resourceNotFoundError",
			err: &resourceNotFoundError{
				resourceType: resources.Channel,
				ref:          "General",
			},
			want: fmt.Sprintf(
				`%s referenced by %q not found`,
				resources.Channel,
				"General",
			),
		},
		{
			name: "resourceEmptyIDError",
			err: &resourceEmptyIDError{
				resourceType: resources.Team,
				ref:          "Team A",
			},
			want: fmt.Sprintf(
				`%s referenced by %q has empty ID`,
				resources.Team,
				"Team A",
			),
		},
		{
			name: "resourceAmbiguousError_two_options",
			err: &resourceAmbiguousError{
				resourceType: resources.Channel,
				ref:          "General",
				options:      []string{"id-1", "id-2"},
			},
			want: fmt.Sprintf(
				"multiple %ss referenced by %q found:\n%s. \nPlease use one of the IDs instead.",
				resources.Channel,
				"General",
				"id-1"+";\n"+"id-2",
			),
		},
		{
			name: "resourceAmbiguousError_empty_options",
			err: &resourceAmbiguousError{
				resourceType: resources.Team,
				ref:          "Team",
				options:      nil,
			},
			want: fmt.Sprintf(
				"multiple %ss referenced by %q found:\n%s. \nPlease use one of the IDs instead.",
				resources.Team,
				"Team",
				strings.Join(nil, ";\n"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}
