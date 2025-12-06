package chats

import (
	"context"

	"github.com/pzsp-teams/lib/internal/api"
)

type Service struct {
	chatAPI api.ChatsAPI
}

func NewService(chatAPI api.ChatsAPI) *Service {
	return &Service{chatAPI: chatAPI}
}

func (s *Service) CreateOneToOne(ctx context.Context, ownerEmail, recipientEmail string) (*Chat, error) {
	resp, err := s.chatAPI.Create(ctx, []string{ownerEmail, recipientEmail}, "")
	if err != nil {
		return nil, mapError(err)
	}
	return mapGraphChat(resp), nil
}

func (s *Service) CreateGroup(ctx context.Context, ownerEmail string, recipientEmails []string, topic string) (*Chat, error) {
	allEmails := make([]string, 0, 1+len(recipientEmails))
	allEmails = append(allEmails, ownerEmail)
	allEmails = append(allEmails, recipientEmails...)
	resp, err := s.chatAPI.Create(ctx, allEmails, topic)
	if err != nil {
		return nil, mapError(err)
	}
	return mapGraphChat(resp), nil
}

func (s *Service) SendMessage(ctx context.Context, chatID, content string) (*ChatMessage, error) {
	resp, err := s.chatAPI.SendMessage(ctx, chatID, content)
	if err != nil {
		return nil, mapError(err)
	}
	return mapGraphChatMessage(resp), nil
}

func (s *Service) ListMyJoined(ctx context.Context) ([]*Chat, error) {
	resp, err := s.chatAPI.ListMyJoined(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	chats := make([]*Chat, len(resp.GetValue()))
	for i, c := range resp.GetValue() {
		chats[i] = mapGraphChat(c)
	}
	return chats, nil
}

func (s *Service) ListMembers(ctx context.Context, chatID string) ([]*ChatMember, error) {
	resp, err := s.chatAPI.ListMembers(ctx, chatID)
	if err != nil {
		return nil, mapError(err)
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
	resp, err := s.chatAPI.AddMember(ctx, chatID, email)
	if err != nil {
		return nil, mapError(err)
	}
	return mapGraphChatMember(resp), nil
}
