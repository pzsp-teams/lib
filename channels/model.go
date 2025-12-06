package channels

import (
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

type MessageContentType string

const (
	MessageContentTypeText MessageContentType = "text"
	MessageContentTypeHTML MessageContentType = "html"
)

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
	ContentType     MessageContentType
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
