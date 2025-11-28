package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// Service will be used later
type Service struct {
	api ChannelAPIInterface
}

// NewService will be used later
func NewService(api ChannelAPIInterface) *Service {
	return &Service{api: api}
}

// ListChannels will be used later
func (s *Service) ListChannels(ctx context.Context, teamID string) ([]*Channel, error) {
	resp, err := s.api.ListChannels(ctx, teamID)
	if err != nil {
		return nil, mapError(err)
	}
	var channels []*Channel
	for _, ch := range resp.GetValue() {
		name := deref(ch.GetDisplayName())
		channels = append(channels, &Channel{
			ID:        deref(ch.GetId()),
			Name:      name,
			IsGeneral: name == "General",
		})
	}
	return channels, nil
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamID, channelID string) (*Channel, error) {
	resp, err := s.api.GetChannel(ctx, teamID, channelID)
	if err != nil {
		return nil, mapError(err)
	}
	name := deref(resp.GetDisplayName())
	return &Channel{
		ID:        deref(resp.GetId()),
		Name:      name,
		IsGeneral: name == "General",
	}, nil
}

// Create will be used later
func (s *Service) Create(ctx context.Context, teamID, name string) (*Channel, error) {
	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)

	created, err := s.api.CreateChannel(ctx, teamID, newChannel)
	if err != nil {
		return nil, mapError(err)
	}
	return &Channel{
		ID:        deref(created.GetId()),
		Name:      deref(created.GetDisplayName()),
		IsGeneral: false,
	}, nil
}

// Delete will be used later
func (s *Service) Delete(ctx context.Context, teamID, channelID string) error {
	err := s.api.DeleteChannel(ctx, teamID, channelID)
	if err != nil {
		return mapError(err)
	}
	return nil
}

// SendMessage will be used later
func (s *Service) SendMessage(ctx context.Context, teamID, channelID string, body MessageBody) (*Message, error) {
	message := msmodels.NewChatMessage()
	messageBody := msmodels.NewItemBody()
	messageBody.SetContent(&body.Content)
	message.SetBody(messageBody)

	resp, err := s.api.SendMessage(ctx, teamID, channelID, message)
	if err != nil {
		return nil, mapError(err)
	}

	return mapChatMessageToMessage(resp), nil
}

// ListMessages retrieves messages from a channel
func (s *Service) ListMessages(ctx context.Context, teamID, channelID string, opts *ListMessagesOptions) ([]*Message, error) {
	var top *int32
	if opts != nil && opts.Top != nil {
		top = opts.Top
	}

	resp, err := s.api.ListMessages(ctx, teamID, channelID, top)
	if err != nil {
		return nil, mapError(err)
	}

	var messages []*Message
	for _, msg := range resp.GetValue() {
		messages = append(messages, mapChatMessageToMessage(msg))
	}

	return messages, nil
}

// GetMessage retrieves a specific message from a channel
func (s *Service) GetMessage(ctx context.Context, teamID, channelID, messageID string) (*Message, error) {
	resp, err := s.api.GetMessage(ctx, teamID, channelID, messageID)
	if err != nil {
		return nil, mapError(err)
	}

	return mapChatMessageToMessage(resp), nil
}

// ListReplies retrieves replies to a specific message
func (s *Service) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) ([]*Message, error) {
	resp, err := s.api.ListReplies(ctx, teamID, channelID, messageID, top)
	if err != nil {
		return nil, mapError(err)
	}

	var replies []*Message
	for _, reply := range resp.GetValue() {
		replies = append(replies, mapChatMessageToMessage(reply))
	}

	return replies, nil
}

// GetReply retrieves a specific reply to a message
func (s *Service) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*Message, error) {
	resp, err := s.api.GetReply(ctx, teamID, channelID, messageID, replyID)
	if err != nil {
		return nil, mapError(err)
	}

	return mapChatMessageToMessage(resp), nil
}

func mapChatMessageToMessage(msg msmodels.ChatMessageable) *Message {
	if msg == nil {
		return nil
	}

	message := &Message{
		ID: deref(msg.GetId()),
	}

	if body := msg.GetBody(); body != nil {
		message.Content = deref(body.GetContent())
		if contentType := body.GetContentType(); contentType != nil {
			message.ContentType = contentType.String()
		}
	}

	if created := msg.GetCreatedDateTime(); created != nil {
		message.CreatedDateTime = *created
	}

	if from := msg.GetFrom(); from != nil {
		if user := from.GetUser(); user != nil {
			message.From = &MessageFrom{
				UserID:      deref(user.GetId()),
				DisplayName: deref(user.GetDisplayName()),
			}
		}
	}

	if replies := msg.GetReplies(); replies != nil {
		message.ReplyCount = len(replies)
	}

	return message
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
