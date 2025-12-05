package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mapper"
)

// Service will be used later
type Service struct {
	api    api.Channels
	mapper mapper.Mapper
}

// NewService will be used later
func NewService(channelsAPI api.Channels, m mapper.Mapper) *Service {
	return &Service{api: channelsAPI, mapper: m}
}

// ListChannels will be used later
func (s *Service) ListChannels(ctx context.Context, teamRef string) ([]*Channel, error) {
	teamID, err := s.mapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	resp, senderErr := s.api.ListChannels(ctx, teamID)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	var chans []*Channel
	for _, ch := range resp.GetValue() {
		name := deref(ch.GetDisplayName())
		chans = append(chans, &Channel{
			ID:        deref(ch.GetId()),
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
	resp, senderErr := s.api.GetChannel(ctx, teamID, channelID)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	name := deref(resp.GetDisplayName())
	return &Channel{
		ID:        deref(resp.GetId()),
		Name:      name,
		IsGeneral: name == "General",
	}, nil
}

// CreateStandardChannel creates a standard channel in a team. All members of the team will have access to the channel.
func (s *Service) CreateStandardChannel(ctx context.Context, teamRef, name string) (*Channel, error) {
	teamID, err := s.mapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)

	created, senderErr := s.api.CreateStandardChannel(ctx, teamID, newChannel)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return &Channel{
		ID:        deref(created.GetId()),
		Name:      deref(created.GetDisplayName()),
		IsGeneral: false,
	}, nil
}

// CreatePrivateChannel creates a private channel in a team with specified members and owners.
func (s *Service) CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*Channel, error) {
	teamID, err := s.mapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	created, senderErr := s.api.CreatePrivateChannelWithMembers(ctx, teamID, name, memberRefs, ownerRefs)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return &Channel{
		ID:        deref(created.GetId()),
		Name:      deref(created.GetDisplayName()),
		IsGeneral: false,
	}, nil
}

// Delete will be used later
func (s *Service) Delete(ctx context.Context, teamRef, channelRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}
	senderErr := s.api.DeleteChannel(ctx, teamID, channelID)
	if senderErr != nil {
		return mapError(senderErr)
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
	messageBody := msmodels.NewItemBody()
	messageBody.SetContent(&body.Content)
	message.SetBody(messageBody)

	resp, senderErr := s.api.SendMessage(ctx, teamID, channelID, message)
	if senderErr != nil {
		return nil, mapError(senderErr)
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

	resp, senderErr := s.api.ListMessages(ctx, teamID, channelID, top)
	if senderErr != nil {
		return nil, mapError(senderErr)
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
	resp, senderErr := s.api.GetMessage(ctx, teamID, channelID, messageID)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return mapChatMessageToMessage(resp), nil
}

// ListReplies retrieves replies to a specific message
func (s *Service) ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32) ([]*Message, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, senderErr := s.api.ListReplies(ctx, teamID, channelID, messageID, top)
	if senderErr != nil {
		return nil, mapError(senderErr)
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
	resp, senderErr := s.api.GetReply(ctx, teamID, channelID, messageID, replyID)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return mapChatMessageToMessage(resp), nil
}

func (s *Service) ListMembers(ctx context.Context, teamRef, channelRef string) ([]*ChannelMember, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	resp, senderErr := s.api.ListMembers(ctx, teamID, channelID)
	if senderErr != nil {
		return nil, mapError(senderErr)
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
	created, senderErr := s.api.AddMember(ctx, teamID, channelID, userRef, role)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return mapConversationMemberToChannelMember(created), nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*ChannelMember, error) {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	memberID, err := s.mapper.MapUserRefToMemberID(ctx, teamID, channelID, userRef)
	if err != nil {
		return nil, err
	}
	role := "member"
	if isOwner {
		role = "owner"
	}
	updated, senderErr := s.api.UpdateMemberRole(ctx, teamID, channelID, memberID, role)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return mapConversationMemberToChannelMember(updated), nil
}

func (s *Service) RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error {
	teamID, channelID, err := s.resolveTeamAndChannelID(ctx, teamRef, channelRef)
	if err != nil {
		return err
	}
	memberID, err := s.mapper.MapUserRefToMemberID(ctx, teamID, channelID, userRef)
	if err != nil {
		return err
	}
	senderErr := s.api.RemoveMember(ctx, teamID, channelID, memberID)
	if senderErr != nil {
		return mapError(senderErr)
	}
	return nil
}

func (s *Service) resolveTeamAndChannelID(ctx context.Context, teamRef, channelRef string) (teamID, channelID string, err error) {
	teamID, err = s.mapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return "", "", err
	}
	channelID, err = s.mapper.MapChannelRefToChannelID(ctx, teamID, channelRef)
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
		ID: deref(msg.GetId()),
	}

	if body := msg.GetBody(); body != nil {
		message.Content = deref(body.GetContent())
		if contentType := body.GetContentType(); contentType != nil {
			message.ContentType = contentType.String()
		}
	}

	if created := msg.GetCreatedDateTime(); created != nil {
		message.CreatedDateTime = *created
	}

	if from := msg.GetFrom(); from != nil {
		if user := from.GetUser(); user != nil {
			message.From = &MessageFrom{
				UserID:      deref(user.GetId()),
				DisplayName: deref(user.GetDisplayName()),
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
		ID:   deref(member.GetId()),
		Role: role,
	}

	if userMember, ok := member.(*msmodels.AadUserConversationMember); ok {
		channelMember.UserID = deref(userMember.GetUserId())
		channelMember.DisplayName = deref(userMember.GetDisplayName())
	}

	return channelMember
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
