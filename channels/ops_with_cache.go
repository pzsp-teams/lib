package channels

import (
	"context"

	"github.com/pzsp-teams/lib/internal/cacher"
	snd "github.com/pzsp-teams/lib/internal/sender"
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

func (o *opsWithCache) ListChannelsByTeamID(ctx context.Context, teamID string) ([]*models.Channel, *snd.RequestError) {
	out, requestErr := o.chanOps.ListChannelsByTeamID(ctx, teamID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	local := util.CopyNonNil(out)
	o.cacheHandler.Runner.Run(func() {
		o.addChannelsToCache(teamID, local...)
	})
	return out, nil
}

func (o *opsWithCache) GetChannelByID(ctx context.Context, teamID, channelID string) (*models.Channel, *snd.RequestError) {
	ch, requestErr := o.chanOps.GetChannelByID(ctx, teamID, channelID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if ch != nil {
		local := *ch
		o.cacheHandler.Runner.Run(func() {
			o.addChannelsToCache(teamID, local)
		})
	}
	return ch, nil
}

func (o *opsWithCache) CreateStandardChannel(ctx context.Context, teamID, name string) (*models.Channel, *snd.RequestError) {
	ch, requestErr := o.chanOps.CreateStandardChannel(ctx, teamID, name)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if ch != nil {
		local := *ch
		o.cacheHandler.Runner.Run(func() {
			o.addChannelsToCache(teamID, local)
		})
	}
	return ch, nil
}

func (o *opsWithCache) CreatePrivateChannel(ctx context.Context, teamID, name string, memberIDs, ownerIDs []string) (*models.Channel, *snd.RequestError) {
	ch, requestErr := o.chanOps.CreatePrivateChannel(ctx, teamID, name, memberIDs, ownerIDs)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if ch != nil {
		local := *ch
		o.cacheHandler.Runner.Run(func() {
			o.addChannelsToCache(teamID, local)
		})
	}
	return ch, nil
}

func (o *opsWithCache) DeleteChannel(ctx context.Context, teamID, channelID, channelRef string) *snd.RequestError {
	requestErr := o.chanOps.DeleteChannel(ctx, teamID, channelID, channelRef)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return requestErr
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeChannelFromCache(teamID, channelRef)
	})
	return nil
}

func (o *opsWithCache) SendMessage(ctx context.Context, teamID, channelID string, body models.MessageBody) (*models.Message, *snd.RequestError) {
	return cacher.WithErrorClear(func() (*models.Message, *snd.RequestError) {
		return o.chanOps.SendMessage(ctx, teamID, channelID, body)
	}, o.cacheHandler)
}

func (o *opsWithCache) SendReply(ctx context.Context, teamID, channelID, messageID string, body models.MessageBody) (*models.Message, *snd.RequestError) {
	return cacher.WithErrorClear(func() (*models.Message, *snd.RequestError) {
		return o.chanOps.SendReply(ctx, teamID, channelID, messageID, body)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMessages(ctx context.Context, teamID, channelID string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError) {
	return cacher.WithErrorClear(func() ([]*models.Message, *snd.RequestError) {
		return o.chanOps.ListMessages(ctx, teamID, channelID, opts)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListReplies(ctx context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError) {
	return cacher.WithErrorClear(func() ([]*models.Message, *snd.RequestError) {
		return o.chanOps.ListReplies(ctx, teamID, channelID, messageID, opts)
	}, o.cacheHandler)
}

func (o *opsWithCache) GetMessage(ctx context.Context, teamID, channelID, messageID string) (*models.Message, *snd.RequestError) {
	return cacher.WithErrorClear(func() (*models.Message, *snd.RequestError) {
		return o.chanOps.GetMessage(ctx, teamID, channelID, messageID)
	}, o.cacheHandler)
}

func (o *opsWithCache) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*models.Message, *snd.RequestError) {
	return cacher.WithErrorClear(func() (*models.Message, *snd.RequestError) {
		return o.chanOps.GetReply(ctx, teamID, channelID, messageID, replyID)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMembers(ctx context.Context, teamID, channelID string) ([]*models.Member, *snd.RequestError) {
	members, requestErr := o.chanOps.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	local := util.CopyNonNil(members)
	o.cacheHandler.Runner.Run(func() {
		o.addMembersToCache(teamID, channelID, local...)
	})
	return members, nil
}

func (o *opsWithCache) AddMember(ctx context.Context, teamID, channelID, userID string, isOwner bool) (*models.Member, *snd.RequestError) {
	member, requestErr := o.chanOps.AddMember(ctx, teamID, channelID, userID, isOwner)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if member != nil {
		local := *member
		o.cacheHandler.Runner.Run(func() {
			o.addMembersToCache(teamID, channelID, local)
		})
	}
	return member, nil
}

func (o *opsWithCache) UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, isOwner bool) (*models.Member, *snd.RequestError) {
	return cacher.WithErrorClear(func() (*models.Member, *snd.RequestError) {
		return o.chanOps.UpdateMemberRoles(ctx, teamID, channelID, memberID, isOwner)
	}, o.cacheHandler)
}

func (o *opsWithCache) RemoveMember(ctx context.Context, teamID, channelID, memberID, userRef string) *snd.RequestError {
	requestErr := o.chanOps.RemoveMember(ctx, teamID, channelID, memberID, userRef)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return requestErr
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




