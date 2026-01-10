package channels

import (
	"context"
	"errors"

	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/search"
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
		return nil, snd.Wrap("ListChannels", err,
			snd.NewParam(resources.TeamRef, teamRef),
		)
	}

	out, err := s.ops.ListChannelsByTeamID(ctx, teamID)
	if err != nil {
		return nil, snd.Wrap("ListChannels", err,
			snd.NewParam(resources.TeamRef, teamRef),
		)
	}

	return out, nil
}

func (s *service) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("Get", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.GetChannelByID(ctx, teamID, channelID)
	if err != nil {
		return nil, snd.Wrap("Get", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, snd.Wrap("CreateStandardChannel", err,
			snd.NewParam(resources.TeamRef, teamRef),
		)
	}

	out, err := s.ops.CreateStandardChannel(ctx, teamID, name)
	if err != nil {
		return nil, snd.Wrap("CreateStandardChannel", err,
			snd.NewParam(resources.TeamRef, teamRef),
		)
	}
	return out, nil
}

func (s *service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, snd.Wrap("CreatePrivateChannel", err,
			snd.NewParam(resources.TeamRef, teamRef),
		)
	}

	out, err := s.ops.CreatePrivateChannel(ctx, teamID, name, memberRefs, ownerRefs)
	if err != nil {
		return nil, snd.Wrap("CreatePrivateChannel", err,
			snd.NewParam(resources.TeamRef, teamRef),
		)
	}
	return out, nil
}

func (s *service) Delete(ctx context.Context, teamRef, channelRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return snd.Wrap("Delete", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return snd.Wrap("Delete", s.ops.DeleteChannel(ctx, teamID, channelID, channelRef),
		snd.NewParam(resources.TeamRef, teamRef),
		snd.NewParam(resources.ChannelRef, channelRef),
	)
}

func (s *service) SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("SendMessage", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.SendMessage(ctx, teamID, channelID, body)
	if err != nil {
		return nil, snd.Wrap("SendMessage", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) SendReply(ctx context.Context, teamRef, channelRef, messageID string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("SendReply", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.SendReply(ctx, teamID, channelID, messageID, body)
	if err != nil {
		return nil, snd.Wrap("SendReply", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions, includeSystem bool, nextLink *string) (*models.MessageCollection, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("ListMessages", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	if nextLink != nil {
		out, err := s.ops.ListMessagesNext(ctx, teamID, channelID, *nextLink, includeSystem)
		if err != nil {
			return nil, snd.Wrap("ListMessages", err,
				snd.NewParam(resources.TeamRef, teamRef),
				snd.NewParam(resources.ChannelRef, channelRef),
			)
		}
		return out, nil
	}
	out, err := s.ops.ListMessages(ctx, teamID, channelID, opts, includeSystem)
	if err != nil {
		return nil, snd.Wrap("ListMessages", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("GetMessage", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.GetMessage(ctx, teamID, channelID, messageID)
	if err != nil {
		return nil, snd.Wrap("GetMessage", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32, includeSystem bool, nextLink *string) (*models.MessageCollection, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("ListReplies", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	if nextLink != nil {
		out, err := s.ops.ListRepliesNext(ctx, teamID, channelID, messageID, *nextLink, includeSystem)
		if err != nil {
			return nil, snd.Wrap("ListReplies", err,
				snd.NewParam(resources.TeamRef, teamRef),
				snd.NewParam(resources.ChannelRef, channelRef),
			)
		}
		return out, nil
	}
	out, err := s.ops.ListReplies(ctx, teamID, channelID, messageID, &models.ListMessagesOptions{Top: top}, includeSystem)
	if err != nil {
		return nil, snd.Wrap("ListReplies", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("GetReply", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.GetReply(ctx, teamID, channelID, messageID, replyID)
	if err != nil {
		return nil, snd.Wrap("GetReply", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	return out, nil
}

func (s *service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("ListMembers", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}

	out, err := s.ops.ListMembers(ctx, teamID, channelID)
	if err != nil {
		return nil, snd.Wrap("ListMembers", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
		)
	}
	return out, nil
}

func (s *service) AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("AddMember", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}
	out, err := s.ops.AddMember(ctx, teamID, channelID, userRef, isOwner)
	if err != nil {
		return nil, snd.Wrap("AddMember", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	return out, nil
}

func (s *service) UpdateMemberRoles(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("UpdateMemberRoles", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.channelResolver.ResolveChannelMemberRefToID(ctx, teamID, channelID, userRef)
	if err != nil {
		return nil, snd.Wrap("UpdateMemberRoles", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	out, err := s.ops.UpdateMemberRoles(ctx, teamID, channelID, memberID, isOwner)
	if err != nil {
		return nil, snd.Wrap("UpdateMemberRoles", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}
	return out, nil
}

func (s *service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return snd.Wrap("RemoveMember", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.channelResolver.ResolveChannelMemberRefToID(ctx, teamID, channelID, userRef)
	if err != nil {
		return snd.Wrap("RemoveMember", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	err = s.ops.RemoveMember(ctx, teamID, channelID, memberID, userRef)
	if err != nil {
		return snd.Wrap("RemoveMember", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.UserRef, userRef),
		)
	}

	return nil
}

func (s *service) GetMentions(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, snd.Wrap("GetMentions", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.MentionRef, rawMentions...),
		)
	}
	out, err := s.ops.GetMentions(ctx, teamID, teamRef, channelRef, channelID, rawMentions)

	if err != nil {
		return nil, snd.Wrap("GetMentions", err,
			snd.NewParam(resources.TeamRef, teamRef),
			snd.NewParam(resources.ChannelRef, channelRef),
			snd.NewParam(resources.MentionRef, rawMentions...),
		)
	}

	return out, nil
}

func (s *service) SearchMessages(ctx context.Context, teamRef, channelRef *string, opts *search.SearchMessagesOptions, cfg *search.SearchConfig) (*search.SearchResults, error) {
	var teamIDptr, channelIDptr *string

	if teamRef != nil {
		teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, *teamRef)
		if err != nil {
			return nil, snd.Wrap("SearchMessages", err, snd.NewParam(resources.TeamRef, *teamRef))
		}
		teamIDptr = &teamID
	}

	if channelRef != nil {
		if teamRef == nil {
			return nil, snd.Wrap("SearchMessages", errors.New("channelRef requires teamRef"),
				snd.NewParam(resources.ChannelRef, *channelRef),
			)
		}
		channelID, err := s.channelResolver.ResolveChannelRefToID(ctx, *teamIDptr, *channelRef)
		if err != nil {
			return nil, snd.Wrap("SearchMessages", err,
				snd.NewParam(resources.TeamRef, *teamRef),
				snd.NewParam(resources.ChannelRef, *channelRef),
			)
		}
		channelIDptr = &channelID
	}

	if cfg == nil {
		cfg = search.DefaultSearchConfig()
	}
	out, err := s.ops.SearchChannelMessages(ctx, teamIDptr, channelIDptr, opts, cfg)
	if err != nil {
		if teamRef == nil {
			return nil, snd.Wrap("SearchMessages", err)
		}
		if channelRef == nil {
			return nil, snd.Wrap("SearchMessages", err, snd.NewParam(resources.TeamRef, *teamRef))
		}
		return nil, snd.Wrap("SearchMessages", err,
			snd.NewParam(resources.TeamRef, *teamRef),
			snd.NewParam(resources.ChannelRef, *channelRef),
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
