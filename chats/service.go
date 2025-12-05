package chats

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
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

func (s *Service) ListMyJoined()

func mapGraphChat(graphChat msmodels.Chatable) *Chat {
	graphMembers := graphChat.GetMembers()
	members := make([]string, len(graphMembers))
	for i, member := range graphMembers {
		members[i] = deref(member.GetDisplayName())
	}
	return &Chat{
		ID:       deref(graphChat.GetId()),
		ChatType: deref(graphChat.GetChatType()).String(),
		Members:  members,
		IsHidden: deref(graphChat.GetIsHiddenForAllMembers()),
		Topic:    graphChat.GetTopic(),
	}
}

func deref[T any](t *T) T {
	var defaultValue T
	if t == nil {
		return defaultValue
	}
	return *t
}
