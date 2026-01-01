package channels

import (
	"context"

	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type serviceWithCache struct {
	svc          Service
	cacheHandler *cacher.CacheHandler
}

func NewServiceWithCache(svc Service, cacheHandler *cacher.CacheHandler) Service {
	if cacheHandler == nil {
		return svc
	}
	return &serviceWithCache{
		svc:          svc,
		cacheHandler: cacheHandler,
	}
}

func (s *serviceWithCache) Wait() {
	s.cacheHandler.Runner.Wait()
}

func (s *serviceWithCache) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	chans, err := s.svc.ListChannels(ctx, teamRef)
	if err != nil {
		s.cacheHandler.OnError()
		return nil, err
	}
	local := util.CopyNonNil(chans)
	s.cacheHandler.Runner.Run(func() {
		s.addChannelsToCache(ctx, local...)
	})
	return chans, nil
}

func (s *serviceWithCache) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	ch, err := s.svc.Get(ctx, teamRef, channelRef)
	if err != nil {
		s.cacheHandler.OnError()
		return nil, err
	}
	if ch != nil {
		local := *ch
		s.cacheHandler.Runner.Run(func() {
			s.addChannelsToCache(ctx, local)
		})
	}
	return ch, nil
}

func (s *serviceWithCache) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	ch, err := s.svc.CreateStandardChannel(ctx, teamRef, name)
	if err != nil {
		s.cacheHandler.OnError()
		return nil, err
	}
	s.updateCacheAfterCreate(ctx, name, ch)
	return ch, nil
}

func (s *serviceWithCache) CreatePrivateChannel(
	ctx context.Context,
	teamRef, name string,
	memberRefs, ownerRefs []string,
) (*models.Channel, error) {
	ch, err := s.svc.CreatePrivateChannel(ctx, teamRef, name, memberRefs, ownerRefs)
	if err != nil {
		s.cacheHandler.OnError()
		return nil, err
	}
	s.updateCacheAfterCreate(ctx, name, ch)
	return ch, nil
}

func (s *serviceWithCache) updateCacheAfterCreate(ctx context.Context, name string, ch *models.Channel) {
	s.cacheHandler.Runner.Run(func() {
		s.removeChannelFromCache(ctx, name)
		if ch == nil {
			return
		}
		local := *ch
		s.addChannelsToCache(ctx, local)
	})
}

func (s *serviceWithCache) Delete(ctx context.Context, teamRef, channelRef string) error {
	if err := s.svc.Delete(ctx, teamRef, channelRef); err != nil {
		s.cacheHandler.OnError()
		return err
	}
	s.cacheHandler.Runner.Run(func() {
		s.removeChannelFromCache(ctx, channelRef)
	})
	return nil
}

func (s *serviceWithCache) SendMessage(
	ctx context.Context,
	teamRef, channelRef string,
	body models.MessageBody,
) (*models.Message, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() (*models.Message, error) {
		return s.svc.SendMessage(ctx, teamRef, channelRef, body)
	})
}

func (s *serviceWithCache) SendReply(
	ctx context.Context,
	teamRef, channelRef, messageID string,
	body models.MessageBody,
) (*models.Message, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() (*models.Message, error) {
		return s.svc.SendReply(ctx, teamRef, channelRef, messageID, body)
	})
}

func (s *serviceWithCache) ListMessages(
	ctx context.Context,
	teamRef, channelRef string,
	opts *models.ListMessagesOptions,
) ([]*models.Message, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() ([]*models.Message, error) {
		return s.svc.ListMessages(ctx, teamRef, channelRef, opts)
	})
}

func (s *serviceWithCache) GetMessage(
	ctx context.Context,
	teamRef, channelRef, messageID string,
) (*models.Message, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() (*models.Message, error) {
		return s.svc.GetMessage(ctx, teamRef, channelRef, messageID)
	})
}

func (s *serviceWithCache) ListReplies(
	ctx context.Context,
	teamRef, channelRef, messageID string,
	top *int32,
) ([]*models.Message, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() ([]*models.Message, error) {
		return s.svc.ListReplies(ctx, teamRef, channelRef, messageID, top)
	})
}

func (s *serviceWithCache) GetReply(
	ctx context.Context,
	teamRef, channelRef, messageID, replyID string,
) (*models.Message, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() (*models.Message, error) {
		return s.svc.GetReply(ctx, teamRef, channelRef, messageID, replyID)
	})
}

func (s *serviceWithCache) ListMembers(
	ctx context.Context,
	teamRef, channelRef string,
) ([]*models.Member, error) {
	members, err := s.svc.ListMembers(ctx, teamRef, channelRef)
	if err != nil {
		s.cacheHandler.OnError()
		return nil, err
	}
	local := util.CopyNonNil(members)
	s.cacheHandler.Runner.Run(func() {
		s.addMembersToCache(ctx, local...)
	})
	return members, nil
}

func (s *serviceWithCache) AddMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*models.Member, error) {
	member, err := s.svc.AddMember(ctx, teamRef, channelRef, userRef, isOwner)
	if err != nil {
		s.cacheHandler.OnError()
		return nil, err
	}
	if member != nil {
		local := *member
		s.cacheHandler.Runner.Run(func() {
			s.addMembersToCache(ctx, local)
		})
	}
	return member, nil
}

func (s *serviceWithCache) UpdateMemberRole(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*models.Member, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() (*models.Member, error) {
		return s.svc.UpdateMemberRole(ctx, teamRef, channelRef, userRef, isOwner)
	})
}

func (s *serviceWithCache) RemoveMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
) error {
	if err := s.svc.RemoveMember(ctx, teamRef, channelRef, userRef); err != nil {
		s.cacheHandler.OnError()
		return err
	}
	s.cacheHandler.Runner.Run(func() {
		s.invalidateMemberCache(ctx, userRef)
	})
	return nil
}

func (s *serviceWithCache) GetMentions(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
	return cacher.WithErrorClear(s.cacheHandler, func() ([]models.Mention, error) {
		return s.svc.GetMentions(ctx, teamRef, channelRef, rawMentions)
	})
}

func (s *serviceWithCache) addChannelsToCache(ctx context.Context, chans ...models.Channel) {
	for _, ch := range chans {
		teamID, ok := ctx.Value(resolvedTeamIDKey).(string)
		if !ok {
			continue
		}
		key := cacher.NewChannelKey(teamID, ch.Name)
		_ = s.cacheHandler.Cacher.Set(key, ch.ID)
	}
}

func (s *serviceWithCache) removeChannelFromCache(ctx context.Context, channelRef string) {
	teamID, ok := ctx.Value(resolvedTeamIDKey).(string)
	if !ok {
		return
	}

	key := cacher.NewChannelKey(teamID, channelRef)
	_ = s.cacheHandler.Cacher.Invalidate(key)
}

func (s *serviceWithCache) addMembersToCache(ctx context.Context, members ...models.Member) {
	for _, member := range members {
		teamID, ok1 := ctx.Value(resolvedTeamIDKey).(string)
		channelID, ok2 := ctx.Value(resolvedChannelIDKey).(string)
		if !ok1 || !ok2 {
			continue
		}

		key := cacher.NewChannelMemberKey(member.Email, teamID, channelID, nil)
		_ = s.cacheHandler.Cacher.Set(key, member.ID)
	}
}

func (s *serviceWithCache) invalidateMemberCache(ctx context.Context, userRef string) {
	teamID, ok1 := ctx.Value(resolvedTeamIDKey).(string)
	channelID, ok2 := ctx.Value(resolvedChannelIDKey).(string)
	if !ok1 || !ok2 {
		return
	}

	key := cacher.NewChannelMemberKey(userRef, teamID, channelID, nil)
	_ = s.cacheHandler.Cacher.Invalidate(key)
}
