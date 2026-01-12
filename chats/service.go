package chats

import (
	"context"
	"fmt"
	"time"

	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/search"
)

type service struct {
	chatOps      chatOps
	chatResolver resolver.ChatResolver
}

// NewService creates a new instance of the chat service.
func NewService(chatOps chatOps, cr resolver.ChatResolver) Service {
	return &service{chatOps: chatOps, chatResolver: cr}
}

func (s *service) CreateOneOnOne(ctx context.Context, recipientRef string) (*models.Chat, error) {
	resp, err := s.chatOps.CreateOneOnOne(ctx, recipientRef)
	if err != nil {
		return nil, snd.Wrap("CreateOneOnOne", err,
			snd.NewParam(resources.UserRef, recipientRef),
		)
	}

	return resp, nil
}

func (s *service) CreateGroup(ctx context.Context, recipientRefs []string, topic string, includeMe bool) (*models.Chat, error) {
	resp, err := s.chatOps.CreateGroup(ctx, recipientRefs, topic, includeMe)
	if err != nil {
		return nil, snd.Wrap("CreateGroup", err,
			snd.NewParam(resources.UserRef, recipientRefs...),
		)
	}

	return resp, nil
}

func (s *service) GetChat(ctx context.Context, chatRef ChatRef) (*models.Chat, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("GetChat", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	var resp *models.Chat
	switch chatRef.(type) {
	case OneOnOneChatRef:
		resp, err = s.chatOps.GetOneOnOneChat(ctx, chatID)
	case GroupChatRef:
		resp, err = s.chatOps.GetGroupChat(ctx, chatID)
	default:
		return nil, snd.Wrap("GetChat", fmt.Errorf("unknown chat reference type"),
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}
	if err != nil {
		return nil, snd.Wrap("GetChat", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) AddMemberToGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) (*models.Member, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("AddMemberToGroupChat", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	resp, err := s.chatOps.AddMemberToGroupChat(ctx, chatID, userRef)
	if err != nil {
		return nil, snd.Wrap("AddMemberToGroupChat", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	return resp, nil
}

func (s *service) RemoveMemberFromGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return snd.Wrap("RemoveMemberFromGroupChat", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.chatResolver.ResolveChatMemberRefToID(ctx, chatID, userRef)
	if err != nil {
		return snd.Wrap("RemoveMemberFromGroupChat", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	err = s.chatOps.RemoveMemberFromGroupChat(ctx, chatID, memberID)
	if err != nil {
		return snd.Wrap("RemoveMemberFromGroupChat", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	return nil
}

func (s *service) ListGroupChatMembers(ctx context.Context, chatRef GroupChatRef) ([]*models.Member, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("ListGroupChatMembers", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
		)
	}

	resp, err := s.chatOps.ListGroupChatMembers(ctx, chatID)
	if err != nil {
		return nil, snd.Wrap("ListGroupChatMembers", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) UpdateGroupChatTopic(ctx context.Context, chatRef GroupChatRef, topic string) (*models.Chat, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("UpdateGroupChatTopic", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
		)
	}

	resp, err := s.chatOps.UpdateGroupChatTopic(ctx, chatID, topic)
	if err != nil {
		return nil, snd.Wrap("UpdateGroupChatTopic", err,
			snd.NewParam(resources.GroupChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) ListMessages(ctx context.Context, chatRef ChatRef, includeSystem bool, nextLink *string) (*models.MessageCollection, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("ListMessages", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}
	if nextLink != nil {
		out, err := s.chatOps.ListMessagesNext(ctx, chatID, *nextLink, includeSystem)
		if err != nil {
			return nil, snd.Wrap("ListMessages", err,
				snd.NewParam(resources.ChatRef, chatRef.get()),
			)
		}
		return out, nil
	}

	resp, err := s.chatOps.ListMessages(ctx, chatID, includeSystem)
	if err != nil {
		return nil, snd.Wrap("ListMessages", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) SendMessage(ctx context.Context, chatRef ChatRef, body models.MessageBody) (*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("SendMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	if err := validateChatMentions(chatRef, body.Mentions); err != nil {
		return nil, snd.Wrap("SendMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	resp, err := s.chatOps.SendMessage(ctx, chatID, body)
	if err != nil {
		return nil, snd.Wrap("SendMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) DeleteMessage(ctx context.Context, chatRef ChatRef, messageID string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return snd.Wrap("DeleteMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	err = s.chatOps.DeleteMessage(ctx, chatID, messageID)
	if err != nil {
		return snd.Wrap("DeleteMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return nil
}

func (s *service) GetMessage(ctx context.Context, chatRef ChatRef, messageID string) (*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("GetMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	resp, err := s.chatOps.GetMessage(ctx, chatID, messageID)
	if err != nil {
		return nil, snd.Wrap("GetMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error) {
	resp, err := s.chatOps.ListChats(ctx, chatType)
	if err != nil {
		return nil, snd.Wrap("ListChats", err)
	}

	return resp, nil
}

func (s *service) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error) {
	resp, err := s.chatOps.ListAllMessages(ctx, startTime, endTime, top)
	if err != nil {
		return nil, snd.Wrap("ListAllMessages", err)
	}

	return resp, nil
}

func (s *service) ListPinnedMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("ListPinnedMessages", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	resp, err := s.chatOps.ListPinnedMessages(ctx, chatID)
	if err != nil {
		return nil, snd.Wrap("ListPinnedMessages", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) PinMessage(ctx context.Context, chatRef ChatRef, messageID string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return snd.Wrap("PinMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	err = s.chatOps.PinMessage(ctx, chatID, messageID)
	if err != nil {
		return snd.Wrap("PinMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return nil
}

func (s *service) UnpinMessage(ctx context.Context, chatRef ChatRef, pinnedMessageID string) error {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return snd.Wrap("UnpinMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	err = s.chatOps.UnpinMessage(ctx, chatID, pinnedMessageID)
	if err != nil {
		return snd.Wrap("UnpinMessage", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return nil
}

func (s *service) GetMentions(ctx context.Context, chatRef ChatRef, rawMentions []string) ([]models.Mention, error) {
	chatID, err := s.resolveChatIDFromRef(ctx, chatRef)
	if err != nil {
		return nil, snd.Wrap("GetMentions", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	isGroup, err := isGroupChatRef(chatRef)
	if err != nil {
		return nil, snd.Wrap("GetMentions", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	resp, err := s.chatOps.GetMentions(ctx, chatID, isGroup, rawMentions)
	if err != nil {
		return nil, snd.Wrap("GetMentions", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
}

func (s *service) SearchMessages(ctx context.Context, chatRef ChatRef, opts *search.SearchMessagesOptions, searchConfig *search.SearchConfig) (*search.SearchResults, error) {
	var chatID *string
	var err error
	if chatRef != nil {
		id, resolveErr := s.resolveChatIDFromRef(ctx, chatRef)
		if resolveErr != nil {
			return nil, snd.Wrap("SearchMessages", resolveErr,
				snd.NewParam(resources.ChatRef, chatRef.get()),
			)
		}
		chatID = &id
	}
	if searchConfig == nil {
		searchConfig = search.DefaultSearchConfig()
	}
	resp, err := s.chatOps.SearchChatMessages(ctx, chatID, opts, searchConfig)
	if err != nil {
		if chatRef == nil {
			return nil, snd.Wrap("SearchMessages", err)
		}
		return nil, snd.Wrap("SearchMessages", err,
			snd.NewParam(resources.ChatRef, chatRef.get()),
		)
	}

	return resp, nil
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
