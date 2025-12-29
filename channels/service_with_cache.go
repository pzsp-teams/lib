package channels

import (
	"context"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type serviceWithCache struct {
	svc             Service
	cache           cacher.Cacher
	teamResolver    resolver.TeamResolver
	channelResolver resolver.ChannelResolver
	runner          util.TaskRunner
}

func newServiceWithCache(
	svc Service,
	cache cacher.Cacher,
	teamResolver resolver.TeamResolver,
	channelResolver resolver.ChannelResolver,
) *serviceWithCache {
	return &serviceWithCache{
		svc:             svc,
		cache:           cache,
		teamResolver:    teamResolver,
		channelResolver: channelResolver,
	}
}

func NewSyncServiceWithCache(svc Service, cache cacher.Cacher, teamResolver resolver.TeamResolver, channelResolver resolver.ChannelResolver) Service {
	newSvc := newServiceWithCache(svc, cache, teamResolver, channelResolver)
	newSvc.runner = &util.SyncRunner{}
	return newSvc
}

func NewAsyncServiceWithCache(svc Service, cache cacher.Cacher, teamResolver resolver.TeamResolver, channelResolver resolver.ChannelResolver) Service {
	newSvc := newServiceWithCache(svc, cache, teamResolver, channelResolver)
	newSvc.runner = &util.AsyncRunner{}
	return newSvc
}

func (s *serviceWithCache) Wait() {
	s.runner.Wait()
}

func (s *serviceWithCache) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	chans, err := s.svc.ListChannels(ctx, teamRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	local := util.CopyNonNil(chans)
	s.runner.Run(func() {
		s.addChannelsToCache(teamRef, local...)
	})
	return chans, nil
}

func (s *serviceWithCache) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	ch, err := s.svc.Get(ctx, teamRef, channelRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	if ch != nil {
		local := *ch
		s.runner.Run(func() {
			s.addChannelsToCache(teamRef, local)
		})
	}
	return ch, nil
}

func (s *serviceWithCache) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	ch, err := s.svc.CreateStandardChannel(ctx, teamRef, name)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.updateCacheAfterCreate(teamRef, name, ch)
	return ch, nil
}

func (s *serviceWithCache) CreatePrivateChannel(
	ctx context.Context,
	teamRef, name string,
	memberRefs, ownerRefs []string,
) (*models.Channel, error) {
	ch, err := s.svc.CreatePrivateChannel(ctx, teamRef, name, memberRefs, ownerRefs)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.updateCacheAfterCreate(teamRef, name, ch)
	return ch, nil
}

func (s *serviceWithCache) updateCacheAfterCreate(teamRef, name string, ch *models.Channel) {
	s.runner.Run(func() {
		s.removeChannelFromCache(teamRef, name)
		if ch == nil {
			return
		}
		local := *ch
		s.addChannelsToCache(teamRef, local)
	})
}

func (s *serviceWithCache) Delete(ctx context.Context, teamRef, channelRef string) error {
	if err := s.svc.Delete(ctx, teamRef, channelRef); err != nil {
		s.onError()
		return err
	}
	s.runner.Run(func() {
		s.removeChannelFromCache(teamRef, channelRef)
	})
	return nil
}

func (s *serviceWithCache) SendMessage(
	ctx context.Context,
	teamRef, channelRef string,
	body models.MessageBody,
) (*models.Message, error) {
	return withErrorClear(func() (*models.Message, error) {
		return s.svc.SendMessage(ctx, teamRef, channelRef, body)
	}, s)
}

func (s *serviceWithCache) ListMessages(
	ctx context.Context,
	teamRef, channelRef string,
	opts *models.ListMessagesOptions,
) ([]*models.Message, error) {
	return withErrorClear(func() ([]*models.Message, error) {
		return s.svc.ListMessages(ctx, teamRef, channelRef, opts)
	}, s)
}

func (s *serviceWithCache) GetMessage(
	ctx context.Context,
	teamRef, channelRef, messageID string,
) (*models.Message, error) {
	return withErrorClear(func() (*models.Message, error) {
		return s.svc.GetMessage(ctx, teamRef, channelRef, messageID)
	}, s)
}

func (s *serviceWithCache) ListReplies(
	ctx context.Context,
	teamRef, channelRef, messageID string,
	top *int32,
) ([]*models.Message, error) {
	return withErrorClear(func() ([]*models.Message, error) {
		return s.svc.ListReplies(ctx, teamRef, channelRef, messageID, top)
	}, s)
}

func (s *serviceWithCache) GetReply(
	ctx context.Context,
	teamRef, channelRef, messageID, replyID string,
) (*models.Message, error) {
	return withErrorClear(func() (*models.Message, error) {
		return s.svc.GetReply(ctx, teamRef, channelRef, messageID, replyID)
	}, s)
}

func (s *serviceWithCache) ListMembers(
	ctx context.Context,
	teamRef, channelRef string,
) ([]*models.Member, error) {
	members, err := s.svc.ListMembers(ctx, teamRef, channelRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	local := util.CopyNonNil(members)
	s.runner.Run(func() {
		s.addMembersToCache(teamRef, channelRef, local...)
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
		s.onError()
		return nil, err
	}
	if member != nil {
		local := *member
		s.runner.Run(func() {
			s.addMembersToCache(teamRef, channelRef, local)
		})
	}
	return member, nil
}

func (s *serviceWithCache) UpdateMemberRole(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*models.Member, error) {
	return withErrorClear(func() (*models.Member, error) {
		return s.svc.UpdateMemberRole(ctx, teamRef, channelRef, userRef, isOwner)
	}, s)
}

func (s *serviceWithCache) RemoveMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
) error {
	if err := s.svc.RemoveMember(ctx, teamRef, channelRef, userRef); err != nil {
		s.onError()
		return err
	}
	s.runner.Run(func() {
		s.invalidateMemberCache(teamRef, channelRef, userRef)
	})
	return nil
}

func (s *serviceWithCache) addChannelsToCache(teamRef string, chans ...models.Channel) {
	if s.cache == nil || len(chans) == 0 {
		return
	}
	ctx := context.Background()
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	for _, ch := range chans {
		name := strings.TrimSpace(ch.Name)
		if name == "" || util.IsLikelyThreadConversationID(name) {
			continue
		}
		key := cacher.NewChannelKey(teamID, name)
		_ = s.cache.Set(key, ch.ID)
	}
}

func (s *serviceWithCache) removeChannelFromCache(teamRef, channelRef string) {
	if s.cache == nil {
		return
	}
	ref := strings.TrimSpace(channelRef)
	if ref == "" || util.IsLikelyThreadConversationID(ref) {
		return
	}
	ctx := context.Background()
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	key := cacher.NewChannelKey(teamID, ref)
	_ = s.cache.Invalidate(key)
}

func (s *serviceWithCache) addMembersToCache(teamRef, channelRef string, members ...models.Member) {
	if s.cache == nil {
		return
	}
	for _, member := range members {
		ref := strings.TrimSpace(strings.TrimSpace(member.Email))
		if ref == "" {
			continue
		}
		ctx := context.Background()
		teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
		if err != nil {
			continue
		}
		channelID, err := s.channelResolver.ResolveChannelRefToID(ctx, teamID, channelRef)
		if err != nil {
			continue
		}
		key := cacher.NewChannelMemberKey(ref, teamID, channelID, nil)
		_ = s.cache.Set(key, member.ID)
	}
}

func (s *serviceWithCache) invalidateMemberCache(teamRef, channelRef, userRef string) {
	if s.cache == nil {
		return
	}
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return
	}

	ctx := context.Background()
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	channelID, err := s.channelResolver.ResolveChannelRefToID(ctx, teamID, channelRef)
	if err != nil {
		return
	}

	key := cacher.NewChannelMemberKey(ref, teamID, channelID, nil)
	_ = s.cache.Invalidate(key)
}

func (s *serviceWithCache) onError() {
	if s.cache == nil {
		return
	}
	s.runner.Run(func() {
		_ = s.cache.Clear()
	})
}

func withErrorClear[T any](
	fn func() (T, error), s *serviceWithCache,
) (T, error) {
	res, err := fn()
	if err != nil {
		s.onError()
		var zero T
		return zero, err
	}
	return res, nil
}
