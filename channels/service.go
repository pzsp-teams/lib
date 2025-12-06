package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mapper"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

// Service will be used later
type Service struct {
	channelAPI    api.ChannelAPI
	teamMapper    mapper.TeamMapper
	channelMapper mapper.ChannelMapper
}

// NewService will be used later
func NewService(channelsAPI api.ChannelAPI, tm mapper.TeamMapper, cm mapper.ChannelMapper) *Service {
	return &Service{channelAPI: channelsAPI, teamMapper: tm, channelMapper: cm}
}

// ListChannels will be used later
func (s *Service) ListChannels(ctx context.Context, teamRef string) ([]*Channel, error) {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.channelAPI.ListChannels(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}
	var chans []*Channel
	for _, ch := range resp.GetValue() {
		name := util.Deref(ch.GetDisplayName())
		chans = append(chans, &Channel{
			ID:        util.Deref(ch.GetId()),
			Name:      name,
			IsGeneral: name == "General",
		})
	}
	return chans, nil
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamRef, channelRef string) (*Channel, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.channelAPI.GetChannel(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}
	name := util.Deref(resp.GetDisplayName())
	return &Channel{
		ID:        util.Deref(resp.GetId()),
		Name:      name,
		IsGeneral: name == "General",
	}, nil
}

// CreateStandardChannel creates a standard channel in a team. All members of the team will have access to the channel.
func (s *Service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*Channel, error) {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)

	created, requestErr := s.channelAPI.CreateStandardChannel(ctx, teamID, newChannel)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}
	return &Channel{
		ID:        util.Deref(created.GetId()),
		Name:      util.Deref(created.GetDisplayName()),
		IsGeneral: false,
	}, nil
}

// CreatePrivateChannel creates a private channel in a team with specified members and owners.
func (s *Service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*Channel, error) {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	created, requestErr := s.channelAPI.CreatePrivateChannelWithMembers(ctx, teamID, name, memberRefs, ownerRefs)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResources(snd.User, append(memberRefs, ownerRefs...)))
	}
	return &Channel{
		ID:        util.Deref(created.GetId()),
		Name:      util.Deref(created.GetDisplayName()),
		IsGeneral: false,
	}, nil
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
func (s *Service) SendMessage(ctx context.Context, teamRef, channelRef string, body MessageBody) (*Message, error) {
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

	return mapChatMessageToMessage(resp), nil
}

// ListMessages retrieves messages from a channel
func (s *Service) ListMessages(ctx context.Context, teamRef, channelRef string, opts *ListMessagesOptions) ([]*Message, error) {
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

	var messages []*Message
	for _, msg := range resp.GetValue() {
		messages = append(messages, mapChatMessageToMessage(msg))
	}

	return messages, nil
}

// GetMessage retrieves a specific message from a channel
func (s *Service) GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.channelAPI.GetMessage(ctx, teamID, channelID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}
	return mapChatMessageToMessage(resp), nil
}

// ListReplies retrieves replies to a specific message
func (s *Service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32) ([]*Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.channelAPI.ListReplies(ctx, teamID, channelID, messageID, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}
	var replies []*Message
	for _, reply := range resp.GetValue() {
		replies = append(replies, mapChatMessageToMessage(reply))
	}

	return replies, nil
}

// GetReply retrieves a specific reply to a message
func (s *Service) GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.channelAPI.GetReply(ctx, teamID, channelID, messageID, replyID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef), snd.WithResource(snd.Message, messageID))
	}
	return mapChatMessageToMessage(resp), nil
}

func (s *Service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*ChannelMember, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.channelAPI.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef), snd.WithResource(snd.Channel, channelRef))
	}
	var members []*ChannelMember
	for _, member := range resp.GetValue() {
		members = append(members, mapConversationMemberToChannelMember(member))
	}
	return members, nil
}

func (s *Service) AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*ChannelMember, error) {
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
	return mapConversationMemberToChannelMember(created), nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*ChannelMember, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	memberID, err := s.channelMapper.MapUserRefToMemberID(ctx, teamID, channelID, userRef)
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
	return mapConversationMemberToChannelMember(updated), nil
}

func (s *Service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}
	memberID, err := s.channelMapper.MapUserRefToMemberID(ctx, teamID, channelID, userRef)
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
	teamID, err = s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return "", "", err
	}
	channelID, err = s.channelMapper.MapChannelRefToChannelID(ctx, teamID, channelRef)
	if err != nil {
		return "", "", err
	}
	return teamID, channelID, nil
}

func mapChatMessageToMessage(msg msmodels.ChatMessageable) *Message {
	if msg == nil {
		return nil
	}
	message := &Message{
		ID: util.Deref(msg.GetId()),
	}
	if body := msg.GetBody(); body != nil {
		message.Content = util.Deref(body.GetContent())

		if ct := body.GetContentType(); ct != nil {
			switch *ct {
			case msmodels.HTML_BODYTYPE:
				message.ContentType = MessageContentTypeHTML
			default:
				message.ContentType = MessageContentTypeText
			}
		} else {
			message.ContentType = MessageContentTypeText
		}
	}
	if created := msg.GetCreatedDateTime(); created != nil {
		message.CreatedDateTime = *created
	}

	if from := msg.GetFrom(); from != nil {
		if user := from.GetUser(); user != nil {
			message.From = &MessageFrom{
				UserID:      util.Deref(user.GetId()),
				DisplayName: util.Deref(user.GetDisplayName()),
			}
		}
	}

	if replies := msg.GetReplies(); replies != nil {
		message.ReplyCount = len(replies)
	}

	return message
}

func mapConversationMemberToChannelMember(member msmodels.ConversationMemberable) *ChannelMember {
	if member == nil {
		return nil
	}
	roles := member.GetRoles()
	role := ""
	if len(roles) > 0 {
		role = roles[0]
	}
	channelMember := &ChannelMember{
		ID:   util.Deref(member.GetId()),
		Role: role,
	}

	if userMember, ok := member.(*msmodels.AadUserConversationMember); ok {
		channelMember.UserID = util.Deref(userMember.GetUserId())
		channelMember.DisplayName = util.Deref(userMember.GetDisplayName())
	}

	return channelMember
}
