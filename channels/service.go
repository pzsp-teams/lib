package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type service struct {
	ops             channelOps
	teamResolver    resolver.TeamResolver
	channelResolver resolver.ChannelResolver
	userAPI         api.UsersAPI
}

// NewService creates a new channels Service instance
func NewService(ops channelOps, tr resolver.TeamResolver, cr resolver.ChannelResolver, userAPI api.UsersAPI) Service {
	return &service{ops: ops, teamResolver: tr, channelResolver: cr, userAPI: userAPI}
}

func (s *service) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.ListChannelsByTeamID(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return out, nil
}

func (s *service) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.GetChannelByID(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef))
	}
	return out, nil
}

func (s *service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.CreateStandardChannel(ctx, teamID, name)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}
	return out, nil
}

func (s *service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.CreatePrivateChannel(ctx, teamID, name, memberRefs, ownerRefs)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}
	return out, nil
}

func (s *service) Delete(ctx context.Context, teamRef, channelRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}

	requestErr := s.ops.DeleteChannel(ctx, teamID, channelID, channelRef)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef))
	}

	return nil
}

func (s *service) SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.SendMessage(ctx, teamID, channelID, body)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef))
	}

	return out, nil
}

func (s *service) SendReply(ctx context.Context, teamRef, channelRef, messageID string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.SendReply(ctx, teamID, channelID, messageID, body)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.Message, messageID))
	}

	return out, nil
}

func (s *service) ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	out, requestErr := s.ops.ListMessages(ctx, teamID, channelID, opts)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef))
	}
	return out, nil
}

func (s *service) GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.GetMessage(ctx, teamID, channelID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.Message, messageID))
	}

	return out, nil
}

func (s *service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.ListReplies(ctx, teamID, channelID, messageID, &models.ListMessagesOptions{Top: top})
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.Message, messageID))
	}
	return out, nil
}

func (s *service) GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.GetReply(ctx, teamID, channelID, messageID, replyID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.Message, messageID))
	}

	return out, nil
}

func (s *service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef))
	}
	return out, nil
}

func (s *service) AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	out, requestErr := s.ops.AddMember(ctx, teamID, channelID, userRef, isOwner)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.User, userRef))
	}

	return out, nil
}

func (s *service) UpdateMemberRoles(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	memberID, err := s.channelResolver.ResolveChannelMemberRefToID(ctx, teamID, channelID, userRef)
	if err != nil {
		return nil, err
	}

	out, requestErr := s.ops.UpdateMemberRoles(ctx, teamID, channelID, memberID, isOwner)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.User, userRef))
	}
	return out, nil
}

func (s *service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}

	memberID, err := s.channelResolver.ResolveChannelMemberRefToID(ctx, teamID, channelID, userRef)
	if err != nil {
		return err
	}

	requestErr := s.ops.RemoveMember(ctx, teamID, channelID, memberID, userRef)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.Channel, channelRef), snd.WithResource(resources.User, userRef))
	}

	return nil
}

func (s *service) GetMentions(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	out := make([]models.Mention, 0, len(rawMentions))
	adder := mentions.NewMentionAdder(&out)

	for _, raw := range rawMentions {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if tryAddTeamOrChannelMention(adder, raw, teamRef, teamID, channelRef, channelID) {
			continue
		}

		if util.IsLikelyEmail(raw) {
			if err := adder.AddUserMention(ctx, raw, s.userAPI); err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("cannot resolve mention reference: %s", raw)
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
