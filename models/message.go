package models

import (
	"time"
)

/*
MessageContentType represents the type of content in a Microsoft Teams message.
It can be either text or HTML.
*/
type MessageContentType string

const (
	// MessageContentTypeText represents plain text content.
	MessageContentTypeText MessageContentType = "text"
	// MessageContentTypeHTML represents HTML content.
	MessageContentTypeHTML MessageContentType = "html"
)

// Message represents a Microsoft Teams chat message. It can be used in both chats and channels.
type Message struct {
	ID              string
	Content         string
	ContentType     MessageContentType
	CreatedDateTime time.Time
	From            *MessageFrom
	ReplyCount      int
}

// MessageFrom represents the sender of a message in Microsoft Teams.
type MessageFrom struct {
	UserID      string
	DisplayName string
}

// MessageBody represents the body of a message in Microsoft Teams.
type MessageBody struct {
	Content     string
	ContentType MessageContentType
	Mentions    []Mention
}

// ListMessagesOptions contains options for listing messages.
type ListMessagesOptions struct {
	Top           *int32
	ExpandReplies bool
}

// MessageCollection represents a collection of messages, potentially with a link to the next page of results.
type MessageCollection struct {
	Messages []*Message
	NextLink *string
}

type SearchPage struct {
	From *int32
	Size *int32
}

// SearchMessagesOptions contains options for searching messages.
type SearchMessagesOptions struct {
	Query       string
	SearchPage  *SearchPage
	From        *string
	IsRead      *bool
	IsMentioned *bool
	To          *string
	StartTime   *time.Time
	EndTime     *time.Time
}

type SearchResult struct {
	Message   *Message
	ChannelID *string
	TeamID    *string
	ChatID    *string
}

type SearchResults struct {
	Messages []*SearchResult
	NextFrom *int32
}

func (s *SearchMessagesOptions) ParseQuery() string {
	query := s.Query

	if s.From != nil {
		query += ` from:"` + *s.From + `"`
	}
	if s.IsRead != nil {
		if *s.IsRead {
			query += ` IsRead:true`
		} else {
			query += ` IsRead:false`
		}
	}
	if s.IsMentioned != nil {
		if *s.IsMentioned {
			query += ` IsMentioned:true`
		} else {
			query += ` IsMentioned:false`
		}
	}
	if s.To != nil {
		query += ` to:"` + *s.To + `"`
	}
	if s.StartTime != nil && s.EndTime != nil {
		query += ` sent:` + s.StartTime.Format(time.RFC3339) + `..` + s.EndTime.Format(time.RFC3339)
		return query
	}
	if s.StartTime != nil {
		query += ` sent>=` + s.StartTime.Format(time.RFC3339)
	}
	if s.EndTime != nil {
		query += ` sent<=` + s.EndTime.Format(time.RFC3339)
	}

	return query
}
