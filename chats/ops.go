package chats

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type ops struct {
	chatAPI api.ChatAPI
	userAPI api.UserAPI
}

func NewOps(chatAPI api.ChatAPI, userAPI api.UserAPI) chatOps {
	return &ops{
		chatAPI: chatAPI,
		userAPI: userAPI,
	}
}

func (o *ops) CreateOneOnOne(ctx context.Context, userID string) (*models.Chat, error) {
	resp, requestErr := o.chatAPI.CreateOneOnOneChat(ctx, userID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.User, userID))
	}
	return adapter.MapGraphChat(resp), nil
}

func (o *ops) CreateGroup(ctx context.Context, userIDs []string, topic string, includeMe bool) (*models.Chat, error) {
	resp, requestErr := o.chatAPI.CreateGroupChat(ctx, userIDs, topic, includeMe)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResources(resources.User, userIDs))
	}
	return adapter.MapGraphChat(resp), nil
}

func (o *ops) AddMemberToGroupChat(ctx context.Context, chatID, userID string) (*models.Member, error) {
	resp, requestErr := o.chatAPI.AddMemberToGroupChat(ctx, chatID, userID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.GroupChat, chatID), snd.WithResource(resources.User, userID))
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) RemoveMemberFromGroupChat(ctx context.Context, chatID, userID string) error {
	return snd.MapError(o.chatAPI.RemoveMemberFromGroupChat(ctx, chatID, userID), snd.WithResource(resources.GroupChat, chatID), snd.WithResource(resources.User, userID))
}

func (o *ops) ListGroupChatMembers(ctx context.Context, chatID string) ([]*models.Member, error) {
	resp, requestErr := o.chatAPI.ListGroupChatMembers(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.GroupChat, chatID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (o *ops) UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (*models.Chat, error) {
	resp, requestErr := o.chatAPI.UpdateGroupChatTopic(ctx, chatID, topic)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.GroupChat, chatID))
	}
	return adapter.MapGraphChat(resp), nil
}

func (o *ops) ListMessages(ctx context.Context, chatID string, includeSystem bool) ([]*models.Message, error) {
	resp, requestErr := o.chatAPI.ListMessages(ctx, chatID, includeSystem)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Chat, chatID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) SendMessage(ctx context.Context, chatID string, body models.MessageBody) (*models.Message, error) {
	ments, err := mentions.PrepareMentions(&body)
	if err != nil {
		return nil, snd.MapError(&snd.RequestError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Failed to prepare mentions: %v", err),
		})
	}
	resp, requestErr := o.chatAPI.SendMessage(ctx, chatID, body.Content, string(body.ContentType), ments)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Chat, chatID))
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) DeleteMessage(ctx context.Context, chatID, messageID string) error {
	return snd.MapError(o.chatAPI.DeleteMessage(ctx, chatID, messageID), snd.WithResource(resources.Chat, chatID), snd.WithResource(resources.Message, messageID))
}

func (o *ops) GetMessage(ctx context.Context, chatID, messageID string) (*models.Message, error) {
	resp, requestErr := o.chatAPI.GetMessage(ctx, chatID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Chat, chatID), snd.WithResource(resources.Message, messageID))
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error) {
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

	resp, requestErr := o.chatAPI.ListChats(ctx, apiType)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphChat), nil
}

func (o *ops) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error) {
	resp, requestErr := o.chatAPI.ListAllMessages(ctx, startTime, endTime, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) ListPinnedMessages(ctx context.Context, chatID string) ([]*models.Message, error) {
	resp, requestErr := o.chatAPI.ListPinnedMessages(ctx, chatID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Chat, chatID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) PinMessage(ctx context.Context, chatID, messageID string) error {
	return snd.MapError(o.chatAPI.PinMessage(ctx, chatID, messageID), snd.WithResource(resources.Chat, chatID), snd.WithResource(resources.Message, messageID))
}

func (o *ops) UnpinMessage(ctx context.Context, chatID, messageID string) error {
	return snd.MapError(o.chatAPI.UnpinMessage(ctx, chatID, messageID), snd.WithResource(resources.Chat, chatID), snd.WithResource(resources.Message, messageID))
}

func (o *ops) GetMentions(ctx context.Context, chatID string, isGroup bool, rawMentions []string) ([]models.Mention, error) {
	out := make([]models.Mention, 0, len(rawMentions))
	adder := mentions.NewMentionAdder(&out)

	for _, raw := range rawMentions {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if ok, err := tryAddEveryoneMention(adder, chatID, isGroup, raw); err != nil {
			return nil, err
		} else if ok {
			continue
		}

		if util.IsLikelyEmail(raw) {
			if err := adder.AddUserMention(ctx, raw, o.userAPI); err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("cannot resolve mention reference: %s", raw)
	}
	return out, nil
}
