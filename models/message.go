package models

import (
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
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
}

func (b MessageBody) ToGraphItemBody() msmodels.ItemBodyable {
	body := msmodels.NewItemBody()
	body.SetContent(&b.Content)
	ct := msmodels.TEXT_BODYTYPE
	if b.ContentType == MessageContentTypeHTML {
		ct = msmodels.HTML_BODYTYPE
	}
	body.SetContentType(&ct)
	return body
}

// ListMessagesOptions contains options for listing messages.
type ListMessagesOptions struct {
	Top           *int32
	ExpandReplies bool
}
