package channels

import (
	"context"

	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type opsWithCache struct {
	chanOps      channelOps
	cacheHandler *cacher.CacheHandler
}

func NewOpsWithCache(chanOps channelOps, cache *cacher.CacheHandler) channelOps {
	if cache == nil {
		return chanOps
	}
	return &opsWithCache{
		chanOps:      chanOps,
		cacheHandler: cache,
	}
}

func (o *opsWithCache) Wait() {
	o.cacheHandler.Runner.Wait()
}

func (o *opsWithCache) ListChannelsByTeamID(ctx context.Context, teamID string) ([]*models.Channel, error) {
	out, err := o.chanOps.ListChannelsByTeamID(ctx, teamID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	local := util.CopyNonNil(out)
	o.cacheHandler.Runner.Run(func() {
		o.addChannelsToCache(teamID, local...)
	})
	return out, nil
}

func (o *opsWithCache) GetChannelByID(ctx context.Context, teamID, channelID string) (*models.Channel, error) {
	ch, err := o.chanOps.GetChannelByID(ctx, teamID, channelID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	if ch != nil {
		local := *ch
		o.cacheHandler.Runner.Run(func() {
			o.addChannelsToCache(teamID, local)
		})
	}
	return ch, nil
}

func (o *opsWithCache) CreateStandardChannel(ctx context.Context, teamID, name string) (*models.Channel, error) {
	ch, err := o.chanOps.CreateStandardChannel(ctx, teamID, name)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	if ch != nil {
		local := *ch
		o.cacheHandler.Runner.Run(func() {
			o.addChannelsToCache(teamID, local)
		})
	}
	return ch, nil
}

func (o *opsWithCache) CreatePrivateChannel(ctx context.Context, teamID, name string, memberIDs, ownerIDs []string) (*models.Channel, error) {
	ch, err := o.chanOps.CreatePrivateChannel(ctx, teamID, name, memberIDs, ownerIDs)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	if ch != nil {
		local := *ch
		o.cacheHandler.Runner.Run(func() {
			o.addChannelsToCache(teamID, local)
		})
	}
	return ch, nil
}

func (o *opsWithCache) DeleteChannel(ctx context.Context, teamID, channelID, channelRef string) error {
	err := o.chanOps.DeleteChannel(ctx, teamID, channelID, channelRef)
	if err != nil {
		o.cacheHandler.OnError(err)
		return err
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeChannelFromCache(teamID, channelRef)
	})
	return nil
}

func (o *opsWithCache) SendMessage(ctx context.Context, teamID, channelID string, body models.MessageBody) (*models.Message, error) {
	return cacher.WithErrorClear(func() (*models.Message, error) {
		return o.chanOps.SendMessage(ctx, teamID, channelID, body)
	}, o.cacheHandler)
}

func (o *opsWithCache) SendReply(ctx context.Context, teamID, channelID, messageID string, body models.MessageBody) (*models.Message, error) {
	return cacher.WithErrorClear(func() (*models.Message, error) {
		return o.chanOps.SendReply(ctx, teamID, channelID, messageID, body)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMessages(ctx context.Context, teamID, channelID string, opts *models.ListMessagesOptions, includeSystem bool) ([]*models.Message, error) {
	return cacher.WithErrorClear(func() ([]*models.Message, error) {
		return o.chanOps.ListMessages(ctx, teamID, channelID, opts, includeSystem)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListReplies(ctx context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions, includeSystem bool) ([]*models.Message, error) {
	return cacher.WithErrorClear(func() ([]*models.Message, error) {
		return o.chanOps.ListReplies(ctx, teamID, channelID, messageID, opts, includeSystem)
	}, o.cacheHandler)
}

func (o *opsWithCache) GetMessage(ctx context.Context, teamID, channelID, messageID string) (*models.Message, error) {
	return cacher.WithErrorClear(func() (*models.Message, error) {
		return o.chanOps.GetMessage(ctx, teamID, channelID, messageID)
	}, o.cacheHandler)
}

func (o *opsWithCache) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*models.Message, error) {
	return cacher.WithErrorClear(func() (*models.Message, error) {
		return o.chanOps.GetReply(ctx, teamID, channelID, messageID, replyID)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMembers(ctx context.Context, teamID, channelID string) ([]*models.Member, error) {
	members, err := o.chanOps.ListMembers(ctx, teamID, channelID)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	local := util.CopyNonNil(members)
	o.cacheHandler.Runner.Run(func() {
		o.addMembersToCache(teamID, channelID, local...)
	})
	return members, nil
}

func (o *opsWithCache) AddMember(ctx context.Context, teamID, channelID, userID string, isOwner bool) (*models.Member, error) {
	member, err := o.chanOps.AddMember(ctx, teamID, channelID, userID, isOwner)
	if err != nil {
		o.cacheHandler.OnError(err)
		return nil, err
	}
	if member != nil {
		local := *member
		o.cacheHandler.Runner.Run(func() {
			o.addMembersToCache(teamID, channelID, local)
		})
	}
	return member, nil
}

func (o *opsWithCache) UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, isOwner bool) (*models.Member, error) {
	return cacher.WithErrorClear(func() (*models.Member, error) {
		return o.chanOps.UpdateMemberRoles(ctx, teamID, channelID, memberID, isOwner)
	}, o.cacheHandler)
}

func (o *opsWithCache) RemoveMember(ctx context.Context, teamID, channelID, memberID, userRef string) error {
	err := o.chanOps.RemoveMember(ctx, teamID, channelID, memberID, userRef)
	if err != nil {
		o.cacheHandler.OnError(err)
		return err
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeMemberFromCache(teamID, channelID, userRef)
	})
	return nil
}

func (o *opsWithCache) addChannelsToCache(teamID string, chans ...models.Channel) {
	if util.AnyBlank(teamID) {
		return
	}
	for _, ch := range chans {
		if util.AnyBlank(ch.Name) {
			continue
		}
		key := cacher.NewChannelKey(teamID, ch.Name)
		_ = o.cacheHandler.Cacher.Set(key, ch.ID)
	}
}

func (o *opsWithCache) removeChannelFromCache(teamID, channelRef string) {
	if util.AnyBlank(teamID, channelRef) {
		return
	}
	key := cacher.NewChannelKey(teamID, channelRef)
	_ = o.cacheHandler.Cacher.Invalidate(key)
}

func (o *opsWithCache) addMembersToCache(teamID, channelID string, members ...models.Member) {
	if util.AnyBlank(teamID, channelID) {
		return
	}
	for _, m := range members {
		if util.AnyBlank(m.Email) {
			continue
		}
		key := cacher.NewChannelMemberKey(teamID, channelID, m.Email, nil)
		_ = o.cacheHandler.Cacher.Set(key, m.ID)
	}
}

func (o *opsWithCache) removeMemberFromCache(teamID, channelID, userRef string) {
	if util.AnyBlank(teamID, channelID, userRef) {
		return
	}
	key := cacher.NewChannelMemberKey(teamID, channelID, userRef, nil)
	_ = o.cacheHandler.Cacher.Invalidate(key)
}

func (o *opsWithCache) GetMentions(ctx context.Context, teamID, teamRef, channelRef, channelID string, rawMentions []string) ([]models.Mention, error) {
	return cacher.WithErrorClear(func() ([]models.Mention, error) {
		return o.chanOps.GetMentions(ctx, teamID, teamRef, channelRef, channelID, rawMentions)
	}, o.cacheHandler)
}
