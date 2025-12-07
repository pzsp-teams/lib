package channels

import (
	"context"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/util"
)

type ServiceWithAutoCacheManagement struct {
	svc   *Service
	cache cacher.Cacher
}

func NewServiceWithAutoCacheManagement(svc *Service, cache cacher.Cacher) *ServiceWithAutoCacheManagement {
	return &ServiceWithAutoCacheManagement{
		svc:   svc,
		cache: cache,
	}
}

func (s *ServiceWithAutoCacheManagement) ListChannels(ctx context.Context, teamRef string) ([]*Channel, error) {
	chans, err := s.svc.ListChannels(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	s.addChannelsToCache(ctx, teamRef, chans)
	return chans, nil
}

func (s *ServiceWithAutoCacheManagement) Get(ctx context.Context, teamRef, channelRef string) (*Channel, error) {
	ch, err := s.svc.Get(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	s.addChannelsToCache(ctx, teamRef, []*Channel{ch})
	return ch, nil
}

func (s *ServiceWithAutoCacheManagement) CreateStandardChannel(ctx context.Context, teamRef, name string) (*Channel, error) {
	ch, err := s.svc.CreateStandardChannel(ctx, teamRef, name)
	if err != nil {
		return nil, err
	}
	s.removeChannelFromCache(ctx, teamRef, name)
	s.addChannelsToCache(ctx, teamRef, []*Channel{ch})
	return ch, nil
}

func (s *ServiceWithAutoCacheManagement) CreatePrivateChannel(
	ctx context.Context,
	teamRef, name string,
	memberRefs, ownerRefs []string,
) (*Channel, error) {
	ch, err := s.svc.CreatePrivateChannel(ctx, teamRef, name, memberRefs, ownerRefs)
	if err != nil {
		return nil, err
	}
	s.removeChannelFromCache(ctx, teamRef, name)
	s.addChannelsToCache(ctx, teamRef, []*Channel{ch})
	return ch, nil
}

func (s *ServiceWithAutoCacheManagement) Delete(ctx context.Context, teamRef, channelRef string) error {
	if err := s.svc.Delete(ctx, teamRef, channelRef); err != nil {
		return err
	}
	s.removeChannelFromCache(ctx, teamRef, channelRef)
	return nil
}

func (s *ServiceWithAutoCacheManagement) SendMessage(
	ctx context.Context,
	teamRef, channelRef string,
	body MessageBody,
) (*Message, error) {
	return s.svc.SendMessage(ctx, teamRef, channelRef, body)
}

func (s *ServiceWithAutoCacheManagement) ListMessages(
	ctx context.Context,
	teamRef, channelRef string,
	opts *ListMessagesOptions,
) ([]*Message, error) {
	return s.svc.ListMessages(ctx, teamRef, channelRef, opts)
}

func (s *ServiceWithAutoCacheManagement) GetMessage(
	ctx context.Context,
	teamRef, channelRef, messageID string,
) (*Message, error) {
	return s.svc.GetMessage(ctx, teamRef, channelRef, messageID)
}

func (s *ServiceWithAutoCacheManagement) ListReplies(
	ctx context.Context,
	teamRef, channelRef, messageID string,
	top *int32,
) ([]*Message, error) {
	return s.svc.ListReplies(ctx, teamRef, channelRef, messageID, top)
}

func (s *ServiceWithAutoCacheManagement) GetReply(
	ctx context.Context,
	teamRef, channelRef, messageID, replyID string,
) (*Message, error) {
	return s.svc.GetReply(ctx, teamRef, channelRef, messageID, replyID)
}

func (s *ServiceWithAutoCacheManagement) ListMembers(
	ctx context.Context,
	teamRef, channelRef string,
) ([]*ChannelMember, error) {
	return s.svc.ListMembers(ctx, teamRef, channelRef)
}

func (s *ServiceWithAutoCacheManagement) AddMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*ChannelMember, error) {
	member, err := s.svc.AddMember(ctx, teamRef, channelRef, userRef, isOwner)
	if err != nil {
		return nil, err
	}
	s.addMemberToCache(ctx, teamRef, channelRef, userRef, member)
	return member, nil
}

func (s *ServiceWithAutoCacheManagement) UpdateMemberRole(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	isOwner bool,
) (*ChannelMember, error) {
	return s.svc.UpdateMemberRole(ctx, teamRef, channelRef, userRef, isOwner)
}

func (s *ServiceWithAutoCacheManagement) RemoveMember(
	ctx context.Context,
	teamRef, channelRef, userRef string,
) error {
	if err := s.svc.RemoveMember(ctx, teamRef, channelRef, userRef); err != nil {
		return err
	}
	s.invalidateMemberCache(ctx, teamRef, channelRef, userRef)
	return nil
}

func (s *ServiceWithAutoCacheManagement) addChannelsToCache(
	ctx context.Context,
	teamRef string,
	chans []*Channel,
) {
	if s.cache == nil || len(chans) == 0 {
		return
	}
	teamID, err := s.svc.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}

	for _, ch := range chans {
		if ch == nil {
			continue
		}
		name := strings.TrimSpace(ch.Name)
		if name == "" || util.IsLikelyChannelID(name) {
			continue
		}
		key := cacher.NewChannelKeyBuilder(teamID, name).ToString()
		_ = s.cache.Set(key, ch.ID)
	}
}

func (s *ServiceWithAutoCacheManagement) removeChannelFromCache(
	ctx context.Context,
	teamRef, channelRef string,
) {
	if s.cache == nil {
		return
	}
	ref := strings.TrimSpace(channelRef)
	if ref == "" || util.IsLikelyChannelID(ref) {
		return
	}
	teamID, err := s.svc.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}
	key := cacher.NewChannelKeyBuilder(teamID, ref).ToString()
	_ = s.cache.Invalidate(key)
}

func (s *ServiceWithAutoCacheManagement) addMemberToCache(
	ctx context.Context,
	teamRef, channelRef, userRef string,
	member *ChannelMember,
) {
	if s.cache == nil || member == nil {
		return
	}
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return
	}

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

func (s *ServiceWithAutoCacheManagement) invalidateMemberCache(
	ctx context.Context,
	teamRef, channelRef, userRef string,
) {
	if s.cache == nil {
		return
	}
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return
	}

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


