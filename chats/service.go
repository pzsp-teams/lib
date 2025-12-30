package chats

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/internal/resolver"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type service struct {
	chatAPI      api.ChatAPI
	chatResolver resolver.ChatResolver
	userAPI      api.UsersAPI
}

// NewService creates a new instance of the chat service.
func NewService(chatAPI api.ChatAPI, cr resolver.ChatResolver, userAPI api.UsersAPI) Service {
	return &service{chatAPI: chatAPI, chatResolver: cr, userAPI: userAPI}
}

func (s *service) CreateOneOneOne(ctx context.Context, recipientRef string) (*models.Chat, error) {
	resp, requestErr := s.chatAPI.CreateOneOnOneChat(ctx, recipientRef)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.User, recipientRef))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *service) CreateGroup(ctx context.Context, recipientRefs []string, topic string, includeMe bool) (*models.Chat, error) {
	resp, requestErr := s.chatAPI.CreateGroupChat(ctx, recipientRefs, topic, includeMe)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(snd.User, recipientRefs))
	}

	return adapter.MapGraphChat(resp), nil
}

func (s *service) AddMemberToGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) (*models.Member, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.AddMemberToGroupChat(ctx, chatID, userRef)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, userRef))
	}

	return adapter.MapGraphMember(resp), nil
}

func (s *service) RemoveMemberFromGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return err
	}

	memberID, err := s.chatResolver.ResolveChatMemberRefToID(ctx, chatID, userRef)
	if err != nil {
		return err
	}

	requestErr := s.chatAPI.RemoveMemberFromGroupChat(ctx, chatID, memberID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID), snd.WithResource(snd.User, userRef))
	}

	return nil
}

func (s *service) ListGroupChatMembers(ctx context.Context, chatRef GroupChatRef) ([]*models.Member, error) {
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

func (s *service) UpdateGroupChatTopic(ctx context.Context, chatRef GroupChatRef, topic string) (*models.Chat, error) {
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

func (s *service) ListMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error) {
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

func (s *service) SendMessage(ctx context.Context, chatRef ChatRef, body models.MessageBody) (*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}
	if len(body.Mentions) > 0 && body.ContentType != models.MessageContentTypeHTML {
		return nil, fmt.Errorf("mentions can only be used with HTML content type")
	}
	if err := validateChatMentions(chatRef, body.Mentions); err != nil {
		return nil, err
	}
	if err := mentions.ValidateAtTags(&body); err != nil {
		return nil, err
	}
	ments, err := mentions.MapMentions(body.Mentions)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.chatAPI.SendMessage(ctx, chatID, body.Content, string(body.ContentType), ments)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Chat, chatID))
	}

	return adapter.MapGraphMessage(resp), nil
}

func (s *service) DeleteMessage(ctx context.Context, chatRef ChatRef, messageID string) error {
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

func (s *service) GetMessage(ctx context.Context, chatRef ChatRef, messageID string) (*models.Message, error) {
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

func (s *service) ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error) {
	var apiType *string

	if chatType != nil {
		switch *chatType {
		case models.ChatTypeGroup:
			groupChat := "group"
			apiType = &groupChat
		case models.ChatTypeOneOnOne:
			oneOnOneChat := "oneOnOne"
			apiType = &oneOnOneChat
		}
	}

	resp, requestErr := s.chatAPI.ListChats(ctx, apiType)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphChat), nil
}

func (s *service) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error) {
	resp, requestErr := s.chatAPI.ListAllMessages(ctx, startTime, endTime, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}

	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (s *service) ListPinnedMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error) {
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

func (s *service) PinMessage(ctx context.Context, chatRef ChatRef, messageID string) error {
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

func (s *service) UnpinMessage(ctx context.Context, chatRef ChatRef, pinnedMessageID string) error {
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

func (s *service) GetMentions(ctx context.Context, chatRef ChatRef, rawMentions []string) ([]models.Mention, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, err
	}

	isGroup, err := isGroupChatRef(chatRef)
	if err != nil {
		return nil, err
	}

	out := make([]models.Mention, 0, len(rawMentions))
	adder := mentions.NewMentionAdder(&out)

	for _, raw := range rawMentions {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if ok, err := s.tryAddEveryoneMention(adder, chatID, isGroup, raw); err != nil {
			return nil, err
		} else if ok {
			continue
		}

		if util.IsLikelyEmail(raw) {
			if err := adder.AddUserMention(ctx, raw, s.userAPI); err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("cannot resolve mention reference: %s", raw)
	}

	return out, nil
}

func isGroupChatRef(chatRef ChatRef) (bool, error) {
	switch chatRef.(type) {
	case GroupChatRef:
		return true, nil
	case OneOnOneChatRef:
		return false, nil
	default:
		return false, fmt.Errorf("unknown chatRef type")
	}
}

func (s *service) tryAddEveryoneMention(adder *mentions.MentionAdder, chatID string, isGroup bool, raw string) (bool, error) {
	low := strings.ToLower(strings.TrimSpace(raw))
	if low != "everyone" && low != "@everyone" {
		return false, nil
	}
	if !isGroup {
		return false, fmt.Errorf("cannot mention everyone in one-on-one chat")
	}
	adder.Add(models.MentionEveryone, chatID, "Everyone", "everyone:"+chatID)
	return true, nil
}

func (s *service) resolveChatIDFromRef(ctx context.Context, chatRef ChatRef) (string, error) {
	switch ref := chatRef.(type) {
	case GroupChatRef:
		return s.chatResolver.ResolveGroupChatRefToID(ctx, ref.Ref)

	case OneOnOneChatRef:
		return s.chatResolver.ResolveOneOnOneChatRefToID(ctx, ref.Ref)

	default:
		return "", fmt.Errorf("unknown chat reference type")
	}
}

func validateChatMentions(chatRef ChatRef, ments []models.Mention) error {
	isOneOnOne := false
	if _, ok := chatRef.(OneOnOneChatRef); ok {
		isOneOnOne = true
	}
	for i := range ments {
		m := ments[i]
		if m.Kind == models.MentionEveryone {
			if isOneOnOne {
				return fmt.Errorf("cannot mention everyone in one-on-one chat")
			}
		}
	}
	return nil
}
