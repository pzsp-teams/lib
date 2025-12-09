package channels

import (
	"context"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type serviceWithAutoCacheManagement struct {
	svc   Service
	cache cacher.Cacher
	teamResolver   resolver.TeamResolver
	channelResolver resolver.ChannelResolver
	run   func(func())
}

func NewSyncServiceWithAutoCacheManagement(svc Service, cache cacher.Cacher, teamResolver resolver.TeamResolver, channelResolver resolver.ChannelResolver) Service {
	return &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: cache,
		teamResolver: teamResolver,
		channelResolver: channelResolver,
		run:   func(fn func()) { fn() },
	}
}

func NewAsyncServiceWithAutoCacheManagement(svc Service, cache cacher.Cacher, teamResolver resolver.TeamResolver, channelResolver resolver.ChannelResolver) Service {
	return &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: cache,
		teamResolver: teamResolver,
		channelResolver: channelResolver,
		run:   func(fn func()) { go fn() },
	}
}

func (s *serviceWithAutoCacheManagement) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	chans, err := s.svc.ListChannels(ctx, teamRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	local := make([]models.Channel, 0, len(chans))
	for _, ch := range chans {
		if ch != nil {
			local = append(local, *ch)
		}
	}
	s.run(func() {
		s.addChannelsToCache(teamRef, local...)
	})
	return chans, nil
}

func (s *serviceWithAutoCacheManagement) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	ch, err := s.svc.Get(ctx, teamRef, channelRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	if ch != nil {
		local := *ch
		s.run(func() {
			s.addChannelsToCache(teamRef, local)
		})
	}
	return ch, nil
}

func (s *serviceWithAutoCacheManagement) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	ch, err := s.svc.CreateStandardChannel(ctx, teamRef, name)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.updateCacheAfterCreate(teamRef, name, ch)
	return ch, nil
}

func (s *serviceWithAutoCacheManagement) CreatePrivateChannel(
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

func (s *serviceWithAutoCacheManagement) updateCacheAfterCreate(teamRef, name string, ch *models.Channel) {
	s.run(func() {
		s.removeChannelFromCache(teamRef, name)
		if ch == nil {
			return
		}
		local := *ch
		s.addChannelsToCache(teamRef, local)
	})
}


func (s *serviceWithAutoCacheManagement) Delete(ctx context.Context, teamRef, channelRef string) error {
	if err := s.svc.Delete(ctx, teamRef, channelRef); err != nil {
		s.onError()
		return err
	}
	s.run(func() {
		s.removeChannelFromCache(teamRef, channelRef)
	})
	return nil
}

func (s *serviceWithAutoCacheManagement) SendMessage(
	ctx context.Context,
	teamRef, channelRef string,
	body models.MessageBody,
) (*models.Message, error) {
	msg, err := s.svc.SendMessage(ctx, teamRef, channelRef, body)
	if err != nil {
		s.onError()
		return nil, err
	}
	return msg, nil
}

func (s *serviceWithAutoCacheManagement) ListMessages(
	ctx context.Context,
	teamRef, channelRef string,
	opts *models.ListMessagesOptions,
) ([]*models.Message, error) {
	msg, err := s.svc.ListMessages(ctx, teamRef, channelRef, opts)
	if err != nil {
		s.onError()
		return nil, err
	}
	return msg, nil
}

func (s *serviceWithAutoCacheManagement) GetMessage(
	ctx context.Context,
	teamRef, channelRef, messageID string,
) (*models.Message, error) {
	msg, err := s.svc.GetMessage(ctx, teamRef, channelRef, messageID)
	if err != nil {
		s.onError()
		return nil, err
	}
	return msg, nil
}

func (s *serviceWithAutoCacheManagement) ListReplies(
	ctx context.Context,
	teamRef, channelRef, messageID string,
	top *int32,
) ([]*models.Message, error) {
	msg, err := s.svc.ListReplies(ctx, teamRef, channelRef, messageID, top)
	if err != nil {
		s.onError()
		return nil, err
	}
	return msg, nil
}

func (s *serviceWithAutoCacheManagement) GetReply(
	ctx context.Context,
	teamRef, channelRef, messageID, replyID string,
) (*models.Message, error) {
	msg, err := s.svc.GetReply(ctx, teamRef, channelRef, messageID, replyID)
	if err != nil {
		s.onError()
		return nil, err
	}
	return msg, nil
}

func (s *serviceWithAutoCacheManagement) ListMembers(
	ctx context.Context,
	teamRef, channelRef string,
) ([]*models.Member, error) {
	members, err := s.svc.ListMembers(ctx, teamRef, channelRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	return members, nil
}

func (s *serviceWithAutoCacheManagement) AddMember(
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
		s.run(func() {
			s.addMemberToCache(teamRef, channelRef, userRef, local)
		})
	}
	return member, nil
}

func (s *serviceWithAutoCacheManagement) UpdateMemberRole(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*models.Member, error) {
	member, err := s.svc.UpdateMemberRole(ctx, teamRef, channelRef, userRef, isOwner)
	if err != nil {
		s.onError()
		return nil, err
	}
	return member, nil
}

func (s *serviceWithAutoCacheManagement) RemoveMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
) error {
	if err := s.svc.RemoveMember(ctx, teamRef, channelRef, userRef); err != nil {
		s.onError()
		return err
	}
	s.run(func() {
		s.invalidateMemberCache(teamRef, channelRef, userRef)
	})
	return nil
}

func (s *serviceWithAutoCacheManagement) addChannelsToCache(
	teamRef string,
	chans ...models.Channel,
) {
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
		if name == "" || util.IsLikelyChannelID(name) {
			continue
		}
		key := cacher.NewChannelKey(teamID, name)
		_ = s.cache.Set(key, ch.ID)
	}
}

func (s *serviceWithAutoCacheManagement) removeChannelFromCache(
	teamRef, channelRef string,
) {
	if s.cache == nil {
		return
	}
	ref := strings.TrimSpace(channelRef)
	if ref == "" || util.IsLikelyChannelID(ref) {
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

func (s *serviceWithAutoCacheManagement) addMemberToCache(
	teamRef, channelRef, userRef string,
	member models.Member,
) {
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

	key := cacher.NewMemberKey(ref, teamID, channelID)
	_ = s.cache.Set(key, member.ID)
}

func (s *serviceWithAutoCacheManagement) invalidateMemberCache(
	teamRef, channelRef, userRef string,
) {
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

	key := cacher.NewMemberKey(ref, teamID, channelID)
	_ = s.cache.Invalidate(key)
}

func (s *serviceWithAutoCacheManagement) onError() {
	if s.cache == nil {
		return
	}
	s.run(func() {
		_ = s.cache.Clear()
	})
}