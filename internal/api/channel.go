package api

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

// ChannesAPI will be used later
type ChannelAPI interface {
	ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError)
	GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError)
	CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError)
	CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberIDs, ownersID []string) (msmodels.Channelable, *sender.RequestError)
	DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError
	SendMessage(ctx context.Context, teamID, channelID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError)
	SendReply(ctx context.Context, teamID, channelID, messageID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError)
	ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError)
	ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError)
	ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError)
	AddMember(ctx context.Context, teamID, channelID, userRef, role string) (msmodels.ConversationMemberable, *sender.RequestError)
	UpdateMemberRole(ctx context.Context, teamID, channelID, memberID, role string) (msmodels.ConversationMemberable, *sender.RequestError)
	RemoveMember(ctx context.Context, teamID, channelID, memberID string) *sender.RequestError
}

type channelAPI struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

// NewChannels will be used later
func NewChannels(client *graph.GraphServiceClient, techParams sender.RequestTechParams) ChannelAPI {
	return &channelAPI{client, techParams}
}

// ListChannels will be used later
func (c *channelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChannelCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChannelCollectionResponseable")
	}

	return out, nil
}

// GetChannel will be used later
func (c *channelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, newTypeError("Channelable")
	}

	return out, nil
}

// CreateStandardChannel creates a standard channel in a team. All members of the team will have access to the channel.
func (c *channelAPI) CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Post(ctx, channel, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, newTypeError("Channelable")
	}

	return out, nil
}

// CreatePrivateChannelWithMembers creates a private channel in a team with specified members and owners.
func (c *channelAPI) CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberRefs, ownerRefs []string) (msmodels.Channelable, *sender.RequestError) {
	ch := msmodels.NewChannel()
	ch.SetDisplayName(&displayName)
	mt := msmodels.PRIVATE_CHANNELMEMBERSHIPTYPE
	ch.SetMembershipType(&mt)

	members := make([]msmodels.ConversationMemberable, 0, len(memberRefs)+len(ownerRefs))
	addToMembers(&members, memberRefs, "member")
	addToMembers(&members, ownerRefs, "owner")
	ch.SetMembers(members)
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Post(ctx, ch, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, newTypeError("Channelable")
	}

	return out, nil
}

// DeleteChannel will be used later
func (c *channelAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		err := c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Delete(ctx, nil)
		return nil, err
	}

	_, err := sender.SendRequest(ctx, call, c.techParams)
	return err
}

// SendMessage will be used later
func (c *channelAPI) SendMessage(ctx context.Context, teamID, channelID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError) {
	message := msmodels.NewChatMessage()
	message.SetBody(messageToGraph(content, contentType))

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			Post(ctx, message, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

// SendReply will be used later
func (c *channelAPI) SendReply(ctx context.Context, teamID, channelID, messageID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError) {
	reply := msmodels.NewChatMessage()
	reply.SetBody(messageToGraph(content, contentType))

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			ByChatMessageId(messageID).
			Replies().
			Post(ctx, reply, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

// ListMessages will be used later
func (c *channelAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		queryParameters := &graphteams.ItemChannelsItemMessagesRequestBuilderGetQueryParameters{}
		if top != nil {
			queryParameters.Top = top
		}
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			Get(ctx, &graphteams.ItemChannelsItemMessagesRequestBuilderGetRequestConfiguration{
				QueryParameters: queryParameters,
			})
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}

	return out, nil
}

// GetMessage will be used later
func (c *channelAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			ByChatMessageId(messageID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

// ListReplies will be used later
func (c *channelAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		queryParameters := &graphteams.ItemChannelsItemMessagesItemRepliesRequestBuilderGetQueryParameters{}
		if top != nil {
			queryParameters.Top = top
		}
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			ByChatMessageId(messageID).
			Replies().
			Get(ctx, &graphteams.ItemChannelsItemMessagesItemRepliesRequestBuilderGetRequestConfiguration{
				QueryParameters: queryParameters,
			})
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}

	return out, nil
}

// GetReply will be used later
func (c *channelAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			ByChatMessageId(messageID).
			Replies().
			ByChatMessageId1(replyID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

func (c *channelAPI) ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Members().
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberCollectionResponseable)
	if !ok {
		return nil, newTypeError("ConversationMemberCollectionResponseable")
	}

	return out, nil
}

func (c *channelAPI) AddMember(ctx context.Context, teamID, channelID, userRef, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
	member := msmodels.NewAadUserConversationMember()
	member.SetRoles([]string{role})
	member.SetAdditionalData(map[string]any{
		graphUserBindKey: fmt.Sprintf(graphUserBindFmt, userRef),
	})
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Members().
			Post(ctx, member, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}

	return out, nil
}

func (c *channelAPI) UpdateMemberRole(ctx context.Context, teamID, channelID, memberID, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
	member := msmodels.NewAadUserConversationMember()
	member.SetRoles([]string{role})
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Members().
			ByConversationMemberId(memberID).
			Patch(ctx, member, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}

	return out, nil
}

func (c *channelAPI) RemoveMember(ctx context.Context, teamID, channelID, memberID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		err := c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Members().
			ByConversationMemberId(memberID).
			Delete(ctx, nil)
		return nil, err
	}

	_, err := sender.SendRequest(ctx, call, c.techParams)
	return err
}
