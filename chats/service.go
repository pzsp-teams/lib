package chats

import (
	"context"

	"github.com/pzsp-teams/lib/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	snd "github.com/pzsp-teams/lib/internal/sender"
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

	chats := make([]*models.Chat, len(resp.GetValue()))
	for i, c := range resp.GetValue() {
		chats[i] = adapter.MapGraphChat(c)
	}

	return chats, nil
}

func (s *Service) ListMembers(ctx context.Context, chatID string) ([]*models.Member, error) {
	resp, requestErr := s.chatAPI.ListMembers(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	members := make([]*models.Member, len(resp.GetValue()))
	for i, m := range resp.GetValue() {
		if result := adapter.MapGraphMember(m); result != nil {
			members[i] = result
		}
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, chatID, email string) (*models.Member, error) {
	resp, requestErr := s.chatAPI.AddMember(ctx, chatID, email)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, email))
	}

	return adapter.MapGraphMember(resp), nil
}
