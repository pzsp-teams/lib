package channels

import (
	"context"

	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type service struct {
	ops             channelOps
	teamResolver    resolver.TeamResolver
	channelResolver resolver.ChannelResolver
}

// NewService creates a new channels Service instance
func NewService(ops channelOps, tr resolver.TeamResolver, cr resolver.ChannelResolver) Service {
	return &service{ops: ops, teamResolver: tr, channelResolver: cr}
}

func (s *service) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("ListChannels", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	out, err := s.ops.ListChannelsByTeamID(ctx, teamID)
	if err != nil {
		return nil, sender.Wrap("ListChannels", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	return out, nil
}

func (s *service) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("Get", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.GetChannelByID(ctx, teamID, channelID)
	if err != nil {
		return nil, sender.Wrap("Get", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("CreateStandardChannel", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	out, err := s.ops.CreateStandardChannel(ctx, teamID, name)
	if err != nil {
		return nil, sender.Wrap("CreateStandardChannel", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}
	return out, nil
}

func (s *service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("CreatePrivateChannel", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	out, err := s.ops.CreatePrivateChannel(ctx, teamID, name, memberRefs, ownerRefs)
	if err != nil {
		return nil, sender.Wrap("CreatePrivateChannel", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}
	return out, nil
}

func (s *service) Delete(ctx context.Context, teamRef, channelRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return sender.Wrap("Delete", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return sender.Wrap("Delete", s.ops.DeleteChannel(ctx, teamID, channelID, channelRef),
		sender.NewParam(resources.TeamRef, teamRef),
		sender.NewParam(resources.ChannelRef, channelRef),
	)
}

func (s *service) SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("SendMessage", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.SendMessage(ctx, teamID, channelID, body)
	if err != nil {
		return nil, sender.Wrap("SendMessage", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) SendReply(ctx context.Context, teamRef, channelRef, messageID string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("SendReply", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.SendReply(ctx, teamID, channelID, messageID, body)
	if err != nil {
		return nil, sender.Wrap("SendReply", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions, includeSystem bool) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("ListMessages", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}
	out, err := s.ops.ListMessages(ctx, teamID, channelID, opts, includeSystem)
	if err != nil {
		return nil, sender.Wrap("ListMessages", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("GetMessage", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.GetMessage(ctx, teamID, channelID, messageID)
	if err != nil {
		return nil, sender.Wrap("GetMessage", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32, includeSystem bool) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("ListReplies", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.ListReplies(ctx, teamID, channelID, messageID, &models.ListMessagesOptions{Top: top}, includeSystem)
	if err != nil {
		return nil, sender.Wrap("ListReplies", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("GetReply", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.GetReply(ctx, teamID, channelID, messageID, replyID)
	if err != nil {
		return nil, sender.Wrap("GetReply", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("ListMembers", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.ListMembers(ctx, teamID, channelID)
	if err != nil {
		return nil, sender.Wrap("ListMembers", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("AddMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}
	out, err := s.ops.AddMember(ctx, teamID, channelID, userRef, isOwner)
	if err != nil {
		return nil, sender.Wrap("AddMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	return out, nil
}

func (s *service) UpdateMemberRoles(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("UpdateMemberRoles", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.channelResolver.ResolveChannelMemberRefToID(ctx, teamID, channelID, userRef)
	if err != nil {
		return nil, sender.Wrap("UpdateMemberRoles", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	out, err := s.ops.UpdateMemberRoles(ctx, teamID, channelID, memberID, isOwner)
	if err != nil {
		return nil, sender.Wrap("UpdateMemberRoles", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}
	return out, nil
}

func (s *service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return sender.Wrap("RemoveMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.channelResolver.ResolveChannelMemberRefToID(ctx, teamID, channelID, userRef)
	if err != nil {
		return sender.Wrap("RemoveMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	err = s.ops.RemoveMember(ctx, teamID, channelID, memberID, userRef)
	if err != nil {
		return sender.Wrap("RemoveMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	return nil
}

func (s *service) GetMentions(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, sender.Wrap("GetMentions", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.MentionRef, rawMentions...),
		)
	}
	out, err := s.ops.GetMentions(ctx, teamID, teamRef, channelRef, channelID, rawMentions)

	if err != nil {
		return nil, sender.Wrap("GetMentions", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.ChannelRef, channelRef),
			sender.NewParam(resources.MentionRef, rawMentions...),
		)
	}

	return out, nil
}

func (s *service) resolveTeamAndChannelID(ctx context.Context, teamRef, channelRef string) (teamID, channelID string, err error) {
	teamID, err = s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return "", "", err
	}
	channelID, err = s.channelResolver.ResolveChannelRefToID(ctx, teamID, channelRef)
	if err != nil {
		return "", "", err
	}
	return teamID, channelID, nil
}
