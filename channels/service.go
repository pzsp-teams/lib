package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resolver"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

// Service will be used later
type Service struct {
	channelAPI      api.ChannelAPI
	teamResolver    resolver.TeamResolver
	channelResolver resolver.ChannelResolver
}

// NewService will be used later
func NewService(channelsAPI api.ChannelAPI, tr resolver.TeamResolver, cr resolver.ChannelResolver) *Service {
	return &Service{channelAPI: channelsAPI, teamResolver: tr, channelResolver: cr}
}

// ListChannels will be used later
func (s *Service) ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.ListChannels(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}

	var chans []*models.Channel
	for _, ch := range resp.GetValue() {
		chans = append(chans, adapter.MapGraphChannel(ch))
	}

	return chans, nil
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
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
func (s *Service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error) {
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
func (s *Service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error) {
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
func (s *Service) Delete(ctx context.Context, teamRef, channelRef string) error {
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
func (s *Service) SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	message := msmodels.NewChatMessage()
	message.SetBody(body.ToGraphItemBody())

	resp, requestErr := s.channelAPI.SendMessage(ctx, teamID, channelID, message)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}

	return adapter.MapGraphMessage(resp), nil
}

// ListMessages retrieves messages from a channel
func (s *Service) ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions) ([]*models.Message, error) {
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

	var messages []*models.Message
	for _, msg := range resp.GetValue() {
		messages = append(messages, adapter.MapGraphMessage(msg))
	}

	return messages, nil
}

// GetMessage retrieves a specific message from a channel
func (s *Service) GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error) {
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
func (s *Service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32) ([]*models.Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.ListReplies(ctx, teamID, channelID, messageID, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}

	var replies []*models.Message
	for _, reply := range resp.GetValue() {
		replies = append(replies, adapter.MapGraphMessage(reply))
	}

	return replies, nil
}

// GetReply retrieves a specific reply to a message
func (s *Service) GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error) {
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

func (s *Service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.channelAPI.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}

	var members []*models.Member
	for _, member := range resp.GetValue() {
		members = append(members, adapter.MapGraphMember(member))
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	role := "member"
	if isOwner {
		role = "owner"
	}

	created, requestErr := s.channelAPI.AddMember(ctx, teamID, channelID, userRef, role)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.User, userRef))
	}

	return adapter.MapGraphMember(created), nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}

	memberID, err := s.channelResolver.ResolveUserRefToMemberID(ctx, teamID, channelID, userRef)
	if err != nil {
		return nil, err
	}

	role := "member"
	if isOwner {
		role = "owner"
	}

	updated, requestErr := s.channelAPI.UpdateMemberRole(ctx, teamID, channelID, memberID, role)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.User, userRef))
	}

	return adapter.MapGraphMember(updated), nil
}

func (s *Service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}

	memberID, err := s.channelResolver.ResolveUserRefToMemberID(ctx, teamID, channelID, userRef)
	if err != nil {
		return err
	}

	requestErr := s.channelAPI.RemoveMember(ctx, teamID, channelID, memberID)
	if requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.User, userRef))
	}

	return nil
}

func (s *Service) resolveTeamAndChannelID(ctx context.Context, teamRef, channelRef string) (teamID, channelID string, err error) {
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
