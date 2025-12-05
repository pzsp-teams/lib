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

func (s *Service) CreateOneToOne(ctx context.Context, ownerEmail, recipientEmail string) (*DirectChat, error) {
	resp, err := s.chatAPI.Create(ctx, []string{ownerEmail, recipientEmail}, "")
	if err != nil {
		return nil, mapError(err)
	}
	return &DirectChat{
		Chat: Chat{
			ID:       deref(resp.GetId()),
			ChatType: (deref(resp.GetChatType())).String(),
			IsHidden: deref(resp.GetIsHiddenForAllMembers()),
		},
	}, nil
}

func (s *Service) CreateGroup(ctx context.Context, ownerEmail string, recipientEmails []string, topic string) (*GroupChat, error) {
	allEmails := make([]string, 0, 1+len(recipientEmails))
	allEmails = append(allEmails, ownerEmail)
	allEmails = append(allEmails, recipientEmails...)
	resp, err := s.chatAPI.Create(ctx, allEmails, topic)
	if err != nil {
		return nil, mapError(err)
	}
	return &GroupChat{
		Chat: Chat{
			ID:       deref(resp.GetId()),
			ChatType: (deref(resp.GetChatType())).String(),
			IsHidden: deref(resp.GetIsHiddenForAllMembers()),
		},
		Topic: topic,
	}, nil
}

func deref[T any](t *T) T {
	var defaultValue T
	if t == nil {
		return defaultValue
	}
	return *t
}
