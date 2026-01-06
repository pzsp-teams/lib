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

// SearchMessagesOptions contains options for searching messages.
type SearchMessagesOptions struct {
	Query           string
	From            *int32
	To              *int32
	Size            *int32
	ExcludeAuthorID *string
	IncludeSystem   bool
	IsRead          *bool
	StartDateTime   *time.Time
	EndDateTime     *time.Time
}
