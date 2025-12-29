package chats

import (
	"context"
	"fmt"
	"time"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resolver"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type Service struct {
	chatAPI      api.ChatAPI
	chatResolver resolver.ChatResolver
}

func NewService(chatAPI api.ChatAPI, chatResolver resolver.ChatResolver) *Service {
	return &Service{chatAPI: chatAPI, chatResolver: chatResolver}
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

func (s *Service) AddMemberToGroupChat(ctx context.Context, chatRef GroupChatRef, memberRef string) (*models.Member, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.AddMemberToGroupChat(ctx, chatID, memberRef)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, memberRef))
	}

	return adapter.MapGraphMember(resp), nil
}

func (s *Service) RemoveMemberFromGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return err
	}

	requestErr := s.chatAPI.RemoveMemberFromGroupChat(ctx, chatID, userRef)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, userRef))
	}

	return nil
}

func (s *Service) ListGroupChatMembers(ctx context.Context, chatRef GroupChatRef) ([]*models.Member, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.ListGroupChatMembers(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (s *Service) UpdateGroupChatTopic(ctx context.Context, chatRef GroupChatRef, topic string) (*models.Chat, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.UpdateGroupChatTopic(ctx, chatID, topic)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *Service) ListMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.ListMessages(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (s *Service) SendMessage(ctx context.Context, chatRef ChatRef, content string, contentType models.MessageContentType) (*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.SendMessage(ctx, chatID, content, string(contentType))
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return adapter.MapGraphMessage(resp), nil
}

func (s *Service) DeleteMessage(ctx context.Context, chatRef ChatRef, messageID string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return err
	}

	requestErr := s.chatAPI.DeleteMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.Message, messageID))
	}

	return nil
}

func (s *Service) GetMessage(ctx context.Context, chatRef ChatRef, messageID string) (*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.GetMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.Message, messageID))
	}

	return adapter.MapGraphMessage(resp), nil
}

func (s *Service) ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error) {
	var apiType string

	if chatType != nil {
		switch *chatType {
		case models.ChatTypeGroup:
			apiType = "group"
		default:
			apiType = "oneOnOne"
		}
	}

	resp, requestErr := s.chatAPI.ListChats(ctx, apiType)
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

func (s *Service) ListPinnedMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.ListPinnedMessages(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (s *Service) PinMessage(ctx context.Context, chatRef ChatRef, messageID string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return err
	}

	requestErr := s.chatAPI.PinMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.Message, messageID))
	}

	return nil
}

func (s *Service) UnpinMessage(ctx context.Context, chatRef ChatRef, pinnedMessageID string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return err
	}

	requestErr := s.chatAPI.UnpinMessage(ctx, chatID, pinnedMessageID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.PinnedMessage, pinnedMessageID))
	}

	return nil
}

func (s *Service) resolveChatIDFromRef(ctx context.Context, chatRef ChatRef) (string, error) {
	switch ref := chatRef.(type) {
	case GroupChatRef:
		return s.chatResolver.ResolveGroupChatRefToID(ctx, ref.Topic)

	case OneOnOneChatRef:
		return s.chatResolver.ResolveOneOnOneChatRefToID(ctx, ref.UserRef)

	case ChatIDRef:
		if ref.ChatID != "" && (util.IsLikelyThreadConversationID(ref.ChatID) || util.IsLikelyChatID(ref.ChatID)) {
			return ref.ChatID, nil
		}
		return "", fmt.Errorf("chat ID %q is not a valid chat identifier", ref.ChatID)

	default:
		return "", fmt.Errorf("unknown chat reference type")
	}
}
