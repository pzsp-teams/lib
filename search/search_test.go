package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func tiPtr(ti TimeInterval) *TimeInterval { return &ti }

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
		in   SearchMessagesOptions
		want string
	}{
		{
			name: "nil query no filters -> empty",
			in:   SearchMessagesOptions{},
			want: "",
		},
		{
			name: "plain query only",
			in: SearchMessagesOptions{
				Query: strPtr("hello world"),
			},
			want: "hello world",
		},
		{
			name: "from single",
			in: SearchMessagesOptions{
				From: []string{"a@x.com"},
			},
			want: ` from:("a@x.com")`,
		},
		{
			name: "from multiple",
			in: SearchMessagesOptions{
				From: []string{"a@x.com", "b@x.com"},
			},
			want: ` from:("a@x.com" OR "b@x.com")`,
		},
		{
			name: "not from multiple",
			in: SearchMessagesOptions{
				NotFrom: []string{"a@x.com", "b@x.com"},
			},
			want: ` NOT from:("a@x.com" OR "b@x.com")`,
		},
		{
			name: "isread true",
			in: SearchMessagesOptions{
				IsRead: boolPtr(true),
			},
			want: ` IsRead:"true"`,
		},
		{
			name: "isread false",
			in: SearchMessagesOptions{
				IsRead: boolPtr(false),
			},
			want: ` IsRead:"false"`,
		},
		{
			name: "ismentioned true",
			in: SearchMessagesOptions{
				IsMentioned: boolPtr(true),
			},
			want: ` IsMentioned:"true"`,
		},
		{
			name: "ismentioned false",
			in: SearchMessagesOptions{
				IsMentioned: boolPtr(false),
			},
			want: ` IsMentioned:"false"`,
		},
		{
			name: "to + notto",
			in: SearchMessagesOptions{
				To:    []string{"x@x.com"},
				NotTo: []string{"y@y.com", "z@z.com"},
			},
			want: ` to:("x@x.com") NOT to:("y@y.com" OR "z@z.com")`,
		},
		{
			name: "interval has priority and returns immediately",
			in: SearchMessagesOptions{
				Query:    strPtr("foo"),
				From:     []string{"a@x.com"},
				Interval: tiPtr(Today),
				StartTime: &start,
				EndTime:   &end,
			},
			want: `foo from:("a@x.com") sent:"today"`,
		},
		{
			name: "start and end -> sent range and returns immediately",
			in: SearchMessagesOptions{
				Query:     strPtr("foo"),
				StartTime: &start,
				EndTime:   &end,
			},
			want: `foo sent:"2026-01-09T00:00:00Z..2026-01-10T23:59:59Z"`,
		},
		{
			name: "only start -> sent>=",
			in: SearchMessagesOptions{
				Query:     strPtr("foo"),
				StartTime: &start,
			},
			want: `foo sent>="2026-01-09T00:00:00Z"`,
		},
		{
			name: "only end -> sent<=",
			in: SearchMessagesOptions{
				Query:   strPtr("foo"),
				EndTime: &end,
			},
			want: `foo sent<="2026-01-10T23:59:59Z"`,
		},
		{
			name: "composition order is stable",
			in: SearchMessagesOptions{
				Query:       strPtr("foo"),
				From:        []string{"a@x.com", "b@x.com"},
				NotFrom:     []string{"c@x.com"},
				IsRead:      boolPtr(true),
				IsMentioned: boolPtr(false),
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
			got := tt.in.ParseQuery()
			require.Equal(t, tt.want, got)
		})
	}
}
