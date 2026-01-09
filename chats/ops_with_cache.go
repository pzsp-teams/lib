package chats

import (
	"context"
	"time"

	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type opsWithCache struct {
	chatOps      chatOps
	cacheHandler *cacher.CacheHandler
}

func NewOpsWithCache(chatOps chatOps, cache *cacher.CacheHandler) chatOps {
	if cache == nil {
		return chatOps
	}
	return &opsWithCache{
		chatOps:      chatOps,
		cacheHandler: cache,
	}
}

func (o *opsWithCache) Wait() {
	o.cacheHandler.Runner.Wait()
}

func (o *opsWithCache) CreateOneOnOne(ctx context.Context, userID string) (*models.Chat, error) {
	chat, err := o.chatOps.CreateOneOnOne(ctx, userID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}

	local := *chat
	o.cacheHandler.Runner.Run(func() {
		o.addChatsToCache(cacheChat{&userID, local})
	})
	return chat, nil
}

func (o *opsWithCache) CreateGroup(ctx context.Context, userIDs []string, topic string, includeMe bool) (*models.Chat, error) {
	chat, err := o.chatOps.CreateGroup(ctx, userIDs, topic, includeMe)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}

	local := *chat
	o.cacheHandler.Runner.Run(func() {
		o.addChatsToCache(cacheChat{nil, local})
	})
	return chat, nil
}

func (o *opsWithCache) AddMemberToGroupChat(ctx context.Context, chatID, userID string) (*models.Member, error) {
	member, err := o.chatOps.AddMemberToGroupChat(ctx, chatID, userID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}

	local := *member
	o.cacheHandler.Runner.Run(func() {
		o.addMembersToCache(chatID, local)
	})

	return member, nil
}

func (o *opsWithCache) RemoveMemberFromGroupChat(ctx context.Context, chatID, userID string) error {
	err := o.chatOps.RemoveMemberFromGroupChat(ctx, chatID, userID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return err
	}

	o.cacheHandler.Runner.Run(func() {
		o.removeMemberFromCache(chatID, userID)
	})

	return nil
}

func (o *opsWithCache) ListGroupChatMembers(ctx context.Context, chatID string) ([]*models.Member, error) {
	members, err := o.chatOps.ListGroupChatMembers(ctx, chatID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}

	local := util.CopyNonNil(members)
	o.cacheHandler.Runner.Run(func() {
		o.addMembersToCache(chatID, local...)
	})

	return members, nil
}

func (o *opsWithCache) UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (*models.Chat, error) {
	return cacher.WithErrorClear(func() (*models.Chat, error) {
		return o.chatOps.UpdateGroupChatTopic(ctx, chatID, topic)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMessages(ctx context.Context, chatID string, includeSystem bool) (*models.MessageCollection, error) {
	return cacher.WithErrorClear(func() (*models.MessageCollection, error) {
		return o.chatOps.ListMessages(ctx, chatID, includeSystem)
	}, o.cacheHandler)
}

func (o *opsWithCache) SendMessage(ctx context.Context, chatID string, body models.MessageBody) (*models.Message, error) {
	return cacher.WithErrorClear(func() (*models.Message, error) {
		return o.chatOps.SendMessage(ctx, chatID, body)
	}, o.cacheHandler)
}

func (o *opsWithCache) DeleteMessage(ctx context.Context, chatID, messageID string) error {
	err := o.chatOps.DeleteMessage(ctx, chatID, messageID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return err
	}
	return nil
}

func (o *opsWithCache) GetMessage(ctx context.Context, chatID, messageID string) (*models.Message, error) {
	return cacher.WithErrorClear(func() (*models.Message, error) {
		return o.chatOps.GetMessage(ctx, chatID, messageID)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error) {
	chats, err := o.chatOps.ListChats(ctx, chatType)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	return chats, nil
}

func (o *opsWithCache) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error) {
	return cacher.WithErrorClear(func() ([]*models.Message, error) {
		return o.chatOps.ListAllMessages(ctx, startTime, endTime, top)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListPinnedMessages(ctx context.Context, chatID string) ([]*models.Message, error) {
	return cacher.WithErrorClear(func() ([]*models.Message, error) {
		return o.chatOps.ListPinnedMessages(ctx, chatID)
	}, o.cacheHandler)
}

func (o *opsWithCache) PinMessage(ctx context.Context, chatID, messageID string) error {
	err := o.chatOps.PinMessage(ctx, chatID, messageID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return err
	}
	return nil
}

func (o *opsWithCache) UnpinMessage(ctx context.Context, chatID, messageID string) error {
	err := o.chatOps.UnpinMessage(ctx, chatID, messageID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return err
	}
	return nil
}

func (o *opsWithCache) GetMentions(ctx context.Context, chatID string, isGroup bool, rawMentions []string) ([]models.Mention, error) {
	return cacher.WithErrorClear(func() ([]models.Mention, error) {
		return o.chatOps.GetMentions(ctx, chatID, isGroup, rawMentions)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMessagesNext(ctx context.Context, chatID, nextLink string, includeSystem bool) (*models.MessageCollection, error) {
	return cacher.WithErrorClear(func() (*models.MessageCollection, error) {
		return o.chatOps.ListMessagesNext(ctx, chatID, nextLink, includeSystem)
	}, o.cacheHandler)
}

type cacheChat struct {
	userID *string
	chat   models.Chat
}

func (o *opsWithCache) addChatsToCache(chats ...cacheChat) {
	for _, item := range chats {
		if item.chat.Type != models.ChatTypeOneOnOne {
			if util.AnyBlank(item.chat.ID, *item.userID) {
				continue
			}
			key := cacher.NewOneOnOneChatKey(*item.userID, nil)
			_ = o.cacheHandler.Cacher.Set(key, item.chat.ID)
		}
		if item.chat.Type == models.ChatTypeGroup {
			if util.AnyBlank(item.chat.ID, util.Deref(item.chat.Topic)) {
				continue
			}
			key := cacher.NewGroupChatKey(*item.chat.Topic)
			_ = o.cacheHandler.Cacher.Set(key, item.chat.ID)
		}
	}
}

func (o *opsWithCache) addMembersToCache(chatID string, members ...models.Member) {
	for _, member := range members {
		if util.AnyBlank(chatID) {
			return
		}
		if util.AnyBlank(member.Email) {
			return
		}
		key := cacher.NewGroupChatMemberKey(chatID, member.Email, nil)
		_ = o.cacheHandler.Cacher.Set(key, member.ID)
	}
}

func (o *opsWithCache) removeMemberFromCache(chatID, userRef string) {
	if util.AnyBlank(chatID, userRef) {
		return
	}
	key := cacher.NewGroupChatMemberKey(chatID, userRef, nil)
	_ = o.cacheHandler.Cacher.Invalidate(key)
}

func (o *opsWithCache) SearchChatMessages(ctx context.Context, chatID *string, opts *models.SearchMessagesOptions) (*models.SearchResults, error) {
	return cacher.WithErrorClear(func() (*models.SearchResults, error) {
		return o.chatOps.SearchChatMessages(ctx, chatID, opts)
	}, o.cacheHandler)
}
