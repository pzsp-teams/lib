package channels

import "time"

// Channel represents a Microsoft Teams channel
type Channel struct {
	ID        string
	Name      string
	IsGeneral bool
}

// Message represents a chat message in a Teams channel
type Message struct {
	ID              string
	Content         string
	ContentType     string
	CreatedDateTime time.Time
	From            *MessageFrom
	ReplyCount      int
}

// MessageFrom represents the sender of a message
type MessageFrom struct {
	UserID      string
	DisplayName string
}

// MessageBody represents the request body for sending a message
type MessageBody struct {
	Content string
}

// ListMessagesOptions contains options for listing messages
type ListMessagesOptions struct {
	Top           *int32
	ExpandReplies bool
}

type ChannelMember struct {
	ID          string
	UserID      string
	DisplayName string
	Role        string
}
