// Package search provides types used to build message-search queries and to represent
// paginated search results returned from the messaging backend (e.g., Microsoft Graph).
//
// The package defines:
//   - query options (SearchMessagesOptions),
//   - pagination controls (SearchPage),
//   - predefined time windows (TimeInterval),
//   - result containers (SearchResult, SearchResults),
//   - and a small concurrency configuration (SearchConfig).
//
// Notes:
//   - If Interval is provided, it takes precedence over StartTime and EndTime.
//   - "To" / "NotTo" filters typically work for chats, not for team channels.
//   - Some providers may not support all filters (e.g., IsRead), depending on API limitations.
package search

import (
	"time"

	"github.com/pzsp-teams/lib/models"
)

// SearchPage defines pagination options for message searches.
//
// From is a zero-based index of the first item to return.
// Size is the maximum number of items to return.
//
// If nil, provider defaults are used (commonly Size=25).
type SearchPage struct {
	From *int32
	Size *int32
}

// TimeInterval is a predefined, human-friendly time window for searching messages.
//
// When Interval is used in SearchMessagesOptions, it overrides StartTime/EndTime.
type TimeInterval string

const (
	// Today selects messages sent today.
	Today TimeInterval = "today"
	// Yesterday selects messages sent yesterday.
	Yesterday TimeInterval = "yesterday"
	// ThisWeek selects messages sent in the current week.
	ThisWeek TimeInterval = "this week"
	// ThisMonth selects messages sent in the current month.
	ThisMonth TimeInterval = "this month"
	// LastMonth selects messages sent in the previous month.
	LastMonth TimeInterval = "last month"
	// ThisYear selects messages sent in the current year.
	ThisYear TimeInterval = "this year"
	// LastYear selects messages sent in the previous year.
	LastYear TimeInterval = "last year"
)

// SearchMessagesOptions describes filters and parameters used to search for messages.
//
// Interval takes precedence over StartTime and EndTime.
//
// Provider notes:
//   - "To" / "NotTo" filters typically work only for chats (not for team channels).
//   - Some providers may not support all filters (e.g., IsRead) due to API limitations.
type SearchMessagesOptions struct {
	// Query is the full-text query string (provider-dependent syntax).
	Query *string

	// SearchPage controls pagination (From/Size).
	SearchPage *SearchPage

	// From includes messages from these sender email addresses.
	From []string
	// NotFrom excludes messages from these sender email addresses.
	NotFrom []string

	// IsRead filters by read status (true=read, false=unread).
	// Note: may be ignored by some providers due to API limitations.
	IsRead *bool

	// IsMentioned filters by whether the current user is mentioned.
	IsMentioned *bool

	// To includes messages addressed to these recipients.
	// Note: commonly chat-only; may not work for channel messages.
	To []string
	// NotTo excludes messages addressed to these recipients.
	// Note: commonly chat-only; may not work for channel messages.
	NotTo []string

	// StartTime is the inclusive start of the sent-time filter.
	StartTime *time.Time
	// EndTime is the exclusive end of the sent-time filter (provider-dependent).
	EndTime *time.Time

	// Interval is a predefined time window; when set it overrides StartTime/EndTime.
	Interval *TimeInterval

	// NotFromMe excludes messages sent by the current user.
	NotFromMe bool
	// NotToMe excludes messages sent to the current user (provider-dependent).
	NotToMe bool
	// FromMe includes only messages sent by the current user.
	FromMe bool
	// ToMe includes only messages sent to the current user (provider-dependent).
	ToMe bool
}

// SearchResult is a single search hit together with its location context.
//
// Exactly which of ChannelID/TeamID/ChatID is set depends on where the message was found.
type SearchResult struct {
	// Message is the found message payload.
	Message *models.Message

	// ChannelID is set when the message was found in a channel.
	ChannelID *string
	// TeamID is set when the message was found in a team (typically with ChannelID).
	TeamID *string
	// ChatID is set when the message was found in a chat.
	ChatID *string
}

// SearchResults is a paginated container of search hits.
type SearchResults struct {
	// Messages contains the list of hits for this page.
	Messages []*SearchResult

	// NextFrom is the pagination cursor/index to continue from (if available).
	NextFrom *int32
}
