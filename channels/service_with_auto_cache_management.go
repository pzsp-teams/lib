package channels

import (
	"context"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type serviceWithAutoCacheManagement struct {
	svc   *service
	cache cacher.Cacher
	run   func(func())
}

func NewServiceWithAutoCacheManagement(svc Service, cache cacher.Cacher) Service {
	return &serviceWithAutoCacheManagement{
		svc:   svc.(*service),
		cache: cache,
		run:   func(fn func()) { go fn() },
	}
}

func (s *serviceWithAutoCacheManagement) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	chans, err := s.svc.ListChannels(ctx, teamRef)
	if err != nil {
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
		return nil, err
	}
	s.updateCacheAfterCreate(teamRef, name, ch)
	return ch, nil
}

func (s *serviceWithAutoCacheManagement) updateCacheAfterCreate(teamRef, name string, ch *models.Channel) {
	if ch != nil {
		local := *ch
		s.run(func() {
			s.removeChannelFromCache(teamRef, name)
			s.addChannelsToCache(teamRef, local)
		})
		return
	}
	s.run(func() {
		s.removeChannelFromCache(teamRef, name)
	})
}

func (s *serviceWithAutoCacheManagement) Delete(ctx context.Context, teamRef, channelRef string) error {
	if err := s.svc.Delete(ctx, teamRef, channelRef); err != nil {
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
	return s.svc.SendMessage(ctx, teamRef, channelRef, body)
}

func (s *serviceWithAutoCacheManagement) ListMessages(
	ctx context.Context,
	teamRef, channelRef string,
	opts *models.ListMessagesOptions,
) ([]*models.Message, error) {
	return s.svc.ListMessages(ctx, teamRef, channelRef, opts)
}

func (s *serviceWithAutoCacheManagement) GetMessage(
	ctx context.Context,
	teamRef, channelRef, messageID string,
) (*models.Message, error) {
	return s.svc.GetMessage(ctx, teamRef, channelRef, messageID)
}

func (s *serviceWithAutoCacheManagement) ListReplies(
	ctx context.Context,
	teamRef, channelRef, messageID string,
	top *int32,
) ([]*models.Message, error) {
	return s.svc.ListReplies(ctx, teamRef, channelRef, messageID, top)
}

func (s *serviceWithAutoCacheManagement) GetReply(
	ctx context.Context,
	teamRef, channelRef, messageID, replyID string,
) (*models.Message, error) {
	return s.svc.GetReply(ctx, teamRef, channelRef, messageID, replyID)
}

func (s *serviceWithAutoCacheManagement) ListMembers(
	ctx context.Context,
	teamRef, channelRef string,
) ([]*models.Member, error) {
	return s.svc.ListMembers(ctx, teamRef, channelRef)
}

func (s *serviceWithAutoCacheManagement) AddMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*models.Member, error) {
	member, err := s.svc.AddMember(ctx, teamRef, channelRef, userRef, isOwner)
	if err != nil {
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
	return s.svc.UpdateMemberRole(ctx, teamRef, channelRef, userRef, isOwner)
}

func (s *serviceWithAutoCacheManagement) RemoveMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
) error {
	if err := s.svc.RemoveMember(ctx, teamRef, channelRef, userRef); err != nil {
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
	teamID, err := s.svc.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	for _, ch := range chans {
		name := strings.TrimSpace(ch.Name)
		if name == "" || util.IsLikelyChannelID(name) {
			continue
		}
		key := cacher.NewChannelKeyBuilder(teamID, name).ToString()
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
	teamID, err := s.svc.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	key := cacher.NewChannelKeyBuilder(teamID, ref).ToString()
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
	teamID, err := s.svc.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	channelID, err := s.svc.channelResolver.ResolveChannelRefToID(ctx, teamID, channelRef)
	if err != nil {
		return
	}

	key := cacher.NewMemberKeyBuilder(ref, teamID, channelID).ToString()
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
	teamID, err := s.svc.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	channelID, err := s.svc.channelResolver.ResolveChannelRefToID(ctx, teamID, channelRef)
	if err != nil {
		return
	}

	key := cacher.NewMemberKeyBuilder(ref, teamID, channelID).ToString()
	_ = s.cache.Invalidate(key)
}
