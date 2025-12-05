package chats

import (
	"context"
	"errors"

	"github.com/pzsp-teams/lib/internal/api"
)

var ErrNotEnoughGuests error = errors.New("you have to specify at least 1 guest to create chat")

type Service struct {
	chatAPI api.ChatsAPI
}

func NewService(chatAPI api.ChatsAPI) *Service {
	return &Service{chatAPI: chatAPI}
}

func (s *Service) CreateChat(ctx context.Context, guestNames []string, guestRole string) (*Chat, error) {
	if len(guestNames) < 1 {
		return nil, ErrNotEnoughGuests
	}
	resp, err := s.chatAPI.CreateChat(ctx, guestNames, guestRole)
	if err != nil {
		return nil, mapError(err)
	}
	return &Chat{
		ID:       deref(resp.GetId()),
		ChatType: (deref(resp.GetChatType())).String(),
		IsHidden: deref(resp.GetIsHiddenForAllMembers()),
	}, nil
}

func deref[T any](t *T) T {
	var defaultValue T
	if t == nil {
		return defaultValue
	}
	return *t
}
