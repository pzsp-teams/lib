package chats

import (
	"context"
	"time"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type Service struct {
	chatAPI api.ChatAPI
}

func NewService(chatAPI api.ChatAPI) *Service {
	return &Service{chatAPI: chatAPI}
}

func (s *Service) CreateOneOneOne(ctx context.Context, recipientRef string) (*models.Chat, error) {
	resp, requestErr := s.chatAPI.CreateOneOnOneChat(ctx, recipientRef)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.User, recipientRef))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *Service) CreateGroup(ctx context.Context, recipientRefs []string, topic string, includeMe bool) (*models.Chat, error) {
	resp, requestErr := s.chatAPI.CreateGroupChat(ctx, recipientRefs, topic, includeMe)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(snd.User, recipientRefs))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *Service) AddMemberToGroupChat(ctx context.Context, chatID, memberRef string) (*models.Member, error) {
	resp, requestErr := s.chatAPI.AddMemberToGroupChat(ctx, chatID, memberRef)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, memberRef))
	}

	return adapter.MapGraphMember(resp), nil
}

func (s *Service) RemoveMemberFromGroupChat(ctx context.Context, chatID, userRef string) error {
	requestErr := s.chatAPI.RemoveMemberFromGroupChat(ctx, chatID, userRef)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, userRef))
	}

	return nil
}

func (s *Service) ListGroupChatMembers(ctx context.Context, chatID string) ([]*models.Member, error) {
	resp, requestErr := s.chatAPI.ListGroupChatMembers(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (s *Service) UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (*models.Chat, error) {
	resp, requestErr := s.chatAPI.UpdateGroupChatTopic(ctx, chatID, topic)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *Service) ListMessages(ctx context.Context, chatID string) ([]*models.Message, error) {
	resp, requestErr := s.chatAPI.ListMessages(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (s *Service) SendMessage(ctx context.Context, chatID, content string, contetType models.MessageContentType) (*models.Message, error) {
	resp, requestErr := s.chatAPI.SendMessage(ctx, chatID, content, string(contetType))
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return adapter.MapGraphMessage(resp), nil
}

func (s *Service) DeleteMessage(ctx context.Context, chatID, messageID string) error {
	requestErr := s.chatAPI.DeleteMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.Message, messageID))
	}

	return nil
}

func (s *Service) GetMessage(ctx context.Context, chatID, messageID string) (*models.Message, error) {
	resp, requestErr := s.chatAPI.GetMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.Message, messageID))
	}

	return adapter.MapGraphMessage(resp), nil
}

func (s *Service) ListMyJoined(ctx context.Context) ([]*models.Chat, error) {
	resp, requestErr := s.chatAPI.ListMyJoined(ctx)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphChat), nil
}

func (s *Service) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error) {
	resp, requestErr := s.chatAPI.ListAllMessages(ctx, startTime, endTime, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (s *Service) ListPinnedMessages(ctx context.Context, chatID string) ([]*models.Message, error) {
	resp, requestErr := s.chatAPI.ListPinnedMessages(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (s *Service) PinMessage(ctx context.Context, chatID, messageID string) error {
	requestErr := s.chatAPI.PinMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.Message, messageID))
	}

	return nil
}

func (s *Service) UnpinMessage(ctx context.Context, chatID, pinnedMessageID string) error {
	requestErr := s.chatAPI.UnpinMessage(ctx, chatID, pinnedMessageID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.PinnedMessage, pinnedMessageID))
	}

	return nil
}
