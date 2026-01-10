package search

import (
	"time"

	"github.com/pzsp-teams/lib/models"
)

// SearchPage contains pagination options for searching messages.
// Fields:
//   - From: The starting index of the search results.
//   - Size: The number of results to return.
//
// Note: If not set, default pagination values will be used. From is zero-based.
// Default Size is typically 25
type SearchPage struct {
	From *int32
	Size *int32
}

// TimeInterval represents a predefined time interval for message searches.
//
// Possible values include:
//   - Today
//   - Yesterday
//   - ThisWeek
//   - ThisMonth
//   - LastMonth
//   - ThisYear
//   - LastYear
type TimeInterval string

const (
	Today     TimeInterval = "today"
	Yesterday TimeInterval = "yesterday"
	ThisWeek  TimeInterval = "this week"
	ThisMonth TimeInterval = "this month"
	LastMonth TimeInterval = "last month"
	ThisYear  TimeInterval = "this year"
	LastYear  TimeInterval = "last year"
)

// SearchMessagesOptions contains options for searching messages.
//
// Fields:
//   - Query: The search query string.
//   - SearchPage: Pagination options.
//   - From: List of sender email addresses to include.
//   - NotFrom: List of sender email addresses to exclude.
//   - IsRead: Filter by read status (true for read, false for unread).
//   - IsMentioned: Filter by mention status (true for mentioned, false for not mentioned).
//   - To: List of recipient email addresses to include.
//   - NotTo: List of recipient email addresses to exclude.
//   - StartTime: Start time for the sent time range filter.
//   - EndTime: End time for the sent time range filter.
//   - Interval: Predefined time interval for the sent time filter.
//   - NotFromMe: Exclude messages sent by the current user.
//   - NotToMe: Exclude messages sent to the current user.
//   - FromMe: Include only messages sent by the current user.
//   - ToMe: Include only messages sent to the current user.
//
// Note: If Interval is set, it takes precedence over StartTime and EndTime.
//
// Note: Using `to` clauses works only in chats, not in team channels.
//
// Note: Currently, the queries for IsRead may not function as expected due to API limitations.
type SearchMessagesOptions struct {
	Query       *string
	SearchPage  *SearchPage
	From        []string
	NotFrom     []string
	IsRead      *bool
	IsMentioned *bool
	To          []string
	NotTo       []string
	StartTime   *time.Time
	EndTime     *time.Time
	Interval    *TimeInterval
	NotFromMe   bool
	NotToMe     bool
	FromMe      bool
	ToMe        bool
}

// SearchResult represents a single search result containing a message and its context.
//
// Fields:
//   - Message: The chat message.
//   - ChannelID: The ID of the channel where the message was found (if applicable).
//   - TeamID: The ID of the team where the message was found (if applicable).
//   - ChatID: The ID of the chat where the message was found (if applicable).
type SearchResult struct {
	Message   *models.Message
	ChannelID *string
	TeamID    *string
	ChatID    *string
}

// SearchResults represents the results of a message search.
//
// Fields:
//   - Messages: A list of search results.
//   - NextFrom: The pagination token for the next page of results (if applicable).
type SearchResults struct {
	Messages []*SearchResult
	NextFrom *int32
}
