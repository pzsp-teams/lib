package api

import (
	"testing"
	"time"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/search"
	"github.com/stretchr/testify/require"
)

func mustTimeRFC3339(t *testing.T, s string) time.Time {
	t.Helper()
	tt, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return tt
}

func TestSearchMessagesOptions_ParseQuery_TableDriven(t *testing.T) {
	t.Parallel()

	start := mustTimeRFC3339(t, "2026-01-09T00:00:00Z")
	end := mustTimeRFC3339(t, "2026-01-10T23:59:59Z")

	tests := []struct {
		name string
		in   search.SearchMessagesOptions
		want string
	}{
		{
			name: "nil query no filters -> empty",
			in:   search.SearchMessagesOptions{},
			want: "",
		},
		{
			name: "plain query only",
			in: search.SearchMessagesOptions{
				Query: util.Ptr("hello world"),
			},
			want: "hello world",
		},
		{
			name: "from single",
			in: search.SearchMessagesOptions{
				From: []string{"a@x.com"},
			},
			want: ` from:("a@x.com")`,
		},
		{
			name: "from multiple",
			in: search.SearchMessagesOptions{
				From: []string{"a@x.com", "b@x.com"},
			},
			want: ` from:("a@x.com" OR "b@x.com")`,
		},
		{
			name: "not from multiple",
			in: search.SearchMessagesOptions{
				NotFrom: []string{"a@x.com", "b@x.com"},
			},
			want: ` NOT from:("a@x.com" OR "b@x.com")`,
		},
		{
			name: "isread true",
			in: search.SearchMessagesOptions{
				IsRead: util.Ptr(true),
			},
			want: ` IsRead:"true"`,
		},
		{
			name: "isread false",
			in: search.SearchMessagesOptions{
				IsRead: util.Ptr(false),
			},
			want: ` IsRead:"false"`,
		},
		{
			name: "ismentioned true",
			in: search.SearchMessagesOptions{
				IsMentioned: util.Ptr(true),
			},
			want: ` IsMentioned:"true"`,
		},
		{
			name: "ismentioned false",
			in: search.SearchMessagesOptions{
				IsMentioned: util.Ptr(false),
			},
			want: ` IsMentioned:"false"`,
		},
		{
			name: "to + notto",
			in: search.SearchMessagesOptions{
				To:    []string{"x@x.com"},
				NotTo: []string{"y@y.com", "z@z.com"},
			},
			want: ` to:("x@x.com") NOT to:("y@y.com" OR "z@z.com")`,
		},
		{
			name: "interval has priority and returns immediately",
			in: search.SearchMessagesOptions{
				Query:     util.Ptr("foo"),
				From:      []string{"a@x.com"},
				Interval:  util.Ptr(search.Today),
				StartTime: &start,
				EndTime:   &end,
			},
			want: `foo from:("a@x.com") sent:"today"`,
		},
		{
			name: "start and end -> sent range and returns immediately",
			in: search.SearchMessagesOptions{
				Query:     util.Ptr("foo"),
				StartTime: &start,
				EndTime:   &end,
			},
			want: `foo sent:"2026-01-09T00:00:00Z..2026-01-10T23:59:59Z"`,
		},
		{
			name: "only start -> sent>=",
			in: search.SearchMessagesOptions{
				Query:     util.Ptr("foo"),
				StartTime: &start,
			},
			want: `foo sent>="2026-01-09T00:00:00Z"`,
		},
		{
			name: "only end -> sent<=",
			in: search.SearchMessagesOptions{
				Query:   util.Ptr("foo"),
				EndTime: &end,
			},
			want: `foo sent<="2026-01-10T23:59:59Z"`,
		},
		{
			name: "composition order is stable",
			in: search.SearchMessagesOptions{
				Query:       util.Ptr("foo"),
				From:        []string{"a@x.com", "b@x.com"},
				NotFrom:     []string{"c@x.com"},
				IsRead:      util.Ptr(true),
				IsMentioned: util.Ptr(false),
				To:          []string{"t@x.com"},
				NotTo:       []string{"nt1@x.com", "nt2@x.com"},
				StartTime:   &start,
				EndTime:     &end,
			},
			want: `foo from:("a@x.com" OR "b@x.com") NOT from:("c@x.com") IsRead:"true" IsMentioned:"false" to:("t@x.com") NOT to:("nt1@x.com" OR "nt2@x.com") sent:"2026-01-09T00:00:00Z..2026-01-10T23:59:59Z"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseQuery(&tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}
