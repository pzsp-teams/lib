package chats

import (
	"context"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type Service struct {
	chatAPI api.ChatsAPI
}

func NewService(chatAPI api.ChatsAPI) *Service {
	return &Service{chatAPI: chatAPI}
}

func (s *Service) CreateOneToOne(ctx context.Context, ownerEmail, recipientEmail string) (*models.Chat, error) {
	allEmails := []string{ownerEmail, recipientEmail}

	resp, requestErr := s.chatAPI.Create(ctx, allEmails, "")
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(snd.User, allEmails))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *Service) CreateGroup(ctx context.Context, ownerEmail string, recipientEmails []string, topic string) (*models.Chat, error) {
	allEmails := append([]string{ownerEmail}, recipientEmails...)

	resp, requestErr := s.chatAPI.Create(ctx, allEmails, topic)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(snd.User, allEmails))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *Service) SendMessage(ctx context.Context, chatID, content string) (*models.Message, error) {
	resp, requestErr := s.chatAPI.SendMessage(ctx, chatID, content)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
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

func (s *Service) ListMembers(ctx context.Context, chatID string) ([]*models.Member, error) {
	resp, requestErr := s.chatAPI.ListMembers(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (s *Service) AddMember(ctx context.Context, chatID, email string) (*models.Member, error) {
	resp, requestErr := s.chatAPI.AddMember(ctx, chatID, email)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, email))
	}

	return adapter.MapGraphMember(resp), nil
}
