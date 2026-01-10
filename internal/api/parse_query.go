package api

import (
	"strings"
	"time"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/search"
)

// ParseQuery constructs the final search query string based on the provided options.
func ParseQuery(s *search.SearchMessagesOptions) string {
	query := util.Deref(s.Query)

	if len(s.From) > 0 {
		query += ` from:("` + strings.Join(s.From, `" OR "`) + `")`
	}
	if len(s.NotFrom) > 0 {
		query += ` NOT from:("` + strings.Join(s.NotFrom, `" OR "`) + `")`
	}
	if s.IsRead != nil {
		if *s.IsRead {
			query += ` IsRead:"true"`
		} else {
			query += ` IsRead:"false"`
		}
	}
	if s.IsMentioned != nil {
		if *s.IsMentioned {
			query += ` IsMentioned:"true"`
		} else {
			query += ` IsMentioned:"false"`
		}
	}
	if len(s.To) > 0 {
		query += ` to:("` + strings.Join(s.To, `" OR "`) + `")`
	}
	if len(s.NotTo) > 0 {
		query += ` NOT to:("` + strings.Join(s.NotTo, `" OR "`) + `")`
	}
	if s.Interval != nil {
		query += ` sent:` + `"` + string(*s.Interval) + `"`
		return query
	}
	if s.StartTime != nil && s.EndTime != nil {
		query += ` sent:` + `"` + s.StartTime.Format(time.RFC3339) + `..` + s.EndTime.Format(time.RFC3339) + `"`
		return query
	}
	if s.StartTime != nil {
		query += ` sent>=` + `"` + s.StartTime.Format(time.RFC3339) + `"`
	}
	if s.EndTime != nil {
		query += ` sent<=` + `"` + s.EndTime.Format(time.RFC3339) + `"`
	}

	return query
}
