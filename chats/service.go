package chats

import (
	"context"

	"github.com/pzsp-teams/lib/internal/api"
	snd "github.com/pzsp-teams/lib/internal/sender"
)

type Service struct {
	chatAPI api.ChatsAPI
}

func NewService(chatAPI api.ChatsAPI) *Service {
	return &Service{chatAPI: chatAPI}
}

func (s *Service) CreateOneToOne(ctx context.Context, ownerEmail, recipientEmail string) (*Chat, error) {
	allEmails := []string{ownerEmail, recipientEmail}
	resp, requestErr := s.chatAPI.Create(ctx, allEmails, "")
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(snd.User, allEmails))
	}
	return mapGraphChat(resp), nil
}

func (s *Service) CreateGroup(ctx context.Context, ownerEmail string, recipientEmails []string, topic string) (*Chat, error) {
	allEmails := append([]string{ownerEmail}, recipientEmails...)
	resp, requestErr := s.chatAPI.Create(ctx, allEmails, topic)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(snd.User, allEmails))
	}
	return mapGraphChat(resp), nil
}

func (s *Service) SendMessage(ctx context.Context, chatID, content string) (*ChatMessage, error) {
	resp, requestErr := s.chatAPI.SendMessage(ctx, chatID, content)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}
	return mapGraphChatMessage(resp), nil
}

func (s *Service) ListMyJoined(ctx context.Context) ([]*Chat, error) {
	resp, requestErr := s.chatAPI.ListMyJoined(ctx)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	chats := make([]*Chat, len(resp.GetValue()))
	for i, c := range resp.GetValue() {
		chats[i] = mapGraphChat(c)
	}
	return chats, nil
}

func (s *Service) ListMembers(ctx context.Context, chatID string) ([]*ChatMember, error) {
	resp, requestErr := s.chatAPI.ListMembers(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}
	members := make([]*ChatMember, len(resp.GetValue()))
	for i, m := range resp.GetValue() {
		if result := mapGraphChatMember(m); result != nil {
			members[i] = result
		}
	}
	return members, nil
}

func (s *Service) AddMember(ctx context.Context, chatID, email string) (*ChatMember, error) {
	resp, requestErr := s.chatAPI.AddMember(ctx, chatID, email)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, email))
	}
	return mapGraphChatMember(resp), nil
}
