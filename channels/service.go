package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resolver"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

// Service will be used later
type service struct {
	channelAPI      api.ChannelAPI
	teamResolver    resolver.TeamResolver
	channelResolver resolver.ChannelResolver
	memberResolver  resolver.MemberResolver
}

// NewService will be used later
func NewService(channelsAPI api.ChannelAPI, tr resolver.TeamResolver, cr resolver.ChannelResolver) Service {
	return &service{channelAPI: channelsAPI, teamResolver: tr, channelResolver: cr}
}

// ListChannels will be used later
func (s *service) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.ListChannels(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphChannel), nil
}

// Get will be used later
func (s *service) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.GetChannel(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}

	return adapter.MapGraphChannel(resp), nil
}

// CreateStandardChannel creates a standard channel in a team. All members of the team will have access to the channel.
func (s *service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)

	created, requestErr := s.channelAPI.CreateStandardChannel(ctx, teamID, newChannel)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}

	return adapter.MapGraphChannel(created), nil
}

// CreatePrivateChannel creates a private channel in a team with specified members and owners.
func (s *service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	created, requestErr := s.channelAPI.CreatePrivateChannelWithMembers(ctx, teamID, name, memberRefs, ownerRefs)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResources(snd.User, append(memberRefs, ownerRefs...)))
	}

	return adapter.MapGraphChannel(created), nil
}

// Delete will be used later
func (s *service) Delete(ctx context.Context, teamRef, channelRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}

	requestErr := s.channelAPI.DeleteChannel(ctx, teamID, channelID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}

	return nil
}

// SendMessage will be used later
func (s *service) SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.SendMessage(ctx, teamID, channelID, body.Content, string(body.ContentType))
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}

	return adapter.MapGraphMessage(resp), nil
}

// ListMessages retrieves messages from a channel
func (s *service) ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	var top *int32
	if opts != nil && opts.Top != nil {
		top = opts.Top
	}
	resp, requestErr := s.channelAPI.ListMessages(ctx, teamID, channelID, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

// GetMessage retrieves a specific message from a channel
func (s *service) GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.GetMessage(ctx, teamID, channelID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}

	return adapter.MapGraphMessage(resp), nil
}

// ListReplies retrieves replies to a specific message
func (s *service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.ListReplies(ctx, teamID, channelID, messageID, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

// GetReply retrieves a specific reply to a message
func (s *service) GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.GetReply(ctx, teamID, channelID, messageID, replyID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}

	return adapter.MapGraphMessage(resp), nil
}

func (s *service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (s *service) AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	role := memberRole(isOwner)
	created, requestErr := s.channelAPI.AddMember(ctx, teamID, channelID, userRef, role)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.User, userRef))
	}

	return adapter.MapGraphMember(created), nil
}

func (s *service) UpdateMemberRole(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	channelMemberCtx := s.memberResolver.NewChannelMemberContext(teamID, channelID, userRef)
	memberID, err := s.memberResolver.ResolveUserRefToMemberID(ctx, channelMemberCtx)
	if err != nil {
		return nil, err
	}

	role := memberRole(isOwner)

	updated, requestErr := s.channelAPI.UpdateMemberRole(ctx, teamID, channelID, memberID, role)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.User, userRef))
	}

	return adapter.MapGraphMember(updated), nil
}

func (s *service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}

	channelMemberCtx := s.memberResolver.NewChannelMemberContext(teamID, channelID, userRef)
	memberID, err := s.memberResolver.ResolveUserRefToMemberID(ctx, channelMemberCtx)
	if err != nil {
		return err
	}

	requestErr := s.channelAPI.RemoveMember(ctx, teamID, channelID, memberID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.User, userRef))
	}

	return nil
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

func memberRole(isOwner bool) string {
	if isOwner {
		return "owner"
	}
	return "member"
}
