// Package api provides interfaces and implementations for Microsoft Teams operations backed by Microsoft Graph.
//
// The package groups functionality into APIs for teams, channels, chats, users, and messaging,
// and returns Graph model types along with request errors from the internal sender layer.
package api

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
)

type ChannelAPI interface {
	ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError)
	GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError)
	CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError)
	CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberIDs, ownersID []string) (msmodels.Channelable, *sender.RequestError)
	DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError
	SendMessage(ctx context.Context, teamID, channelID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError)
	SendReply(ctx context.Context, teamID, channelID, messageID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError)
	ListMessages(ctx context.Context, teamID, channelID string, top *int32, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError)
	ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError)
	ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError)
	AddMember(ctx context.Context, teamID, channelID, userRef string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError)
	UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError)
	RemoveMember(ctx context.Context, teamID, channelID, memberID string) *sender.RequestError
	ListMessagesNext(ctx context.Context, teamID, channelID, nextLink string, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	ListRepliesNext(ctx context.Context, teamID, channelID, messageID, nextLink string, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
}

type channelAPI struct {
	client    *graph.GraphServiceClient
	senderCfg *config.SenderConfig
}

func NewChannels(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) ChannelAPI {
	return &channelAPI{client, senderCfg}
}

func (c *channelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChannelCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChannelCollectionResponseable")
	}

	return out, nil
}

func (c *channelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, newTypeError("Channelable")
	}

	return out, nil
}

func (c *channelAPI) CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Post(ctx, channel, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, newTypeError("Channelable")
	}

	return out, nil
}

func (c *channelAPI) CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberRefs, ownerRefs []string) (msmodels.Channelable, *sender.RequestError) {
	ch := msmodels.NewChannel()
	ch.SetDisplayName(&displayName)
	mt := msmodels.PRIVATE_CHANNELMEMBERSHIPTYPE
	ch.SetMembershipType(&mt)

	members := make([]msmodels.ConversationMemberable, 0, len(memberRefs)+len(ownerRefs))
	addToMembers(&members, memberRefs, []string{})
	addToMembers(&members, ownerRefs, []string{roleOwner})
	ch.SetMembers(members)
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Post(ctx, ch, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, newTypeError("Channelable")
	}

	return out, nil
}

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

	_, err := sender.SendRequest(ctx, call, c.senderCfg)
	return err
}

func (c *channelAPI) SendMessage(ctx context.Context, teamID, channelID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
	message := msmodels.NewChatMessage()
	message.SetBody(messageToGraph(content, contentType))
	if len(mentions) > 0 {
		message.SetMentions(mentions)
	}

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			Post(ctx, message, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

func (c *channelAPI) SendReply(ctx context.Context, teamID, channelID, messageID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
	reply := msmodels.NewChatMessage()
	reply.SetBody(messageToGraph(content, contentType))
	if len(mentions) > 0 {
		reply.SetMentions(mentions)
	}

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

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

func (c *channelAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
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

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}
	if !includeSystem {
		filtered := filterOutSystemEvents(out)
		out.SetValue(filtered)
	}

	return out, nil
}

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

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

func (c *channelAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
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

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}

	if !includeSystem {
		filtered := filterOutSystemEvents(out)
		out.SetValue(filtered)
	}

	return out, nil
}

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

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
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

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberCollectionResponseable)
	if !ok {
		return nil, newTypeError("ConversationMemberCollectionResponseable")
	}

	return out, nil
}

// Roles must be ["owner"] or [] (member)
func (c *channelAPI) AddMember(ctx context.Context, teamID, channelID, userRef string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
	member := newAadUserMemberBody(userRef, roles)
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Members().
			Post(ctx, member, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}

	return out, nil
}

func (c *channelAPI) UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
	patch := newRolesPatchBody(roles)
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Members().
			ByConversationMemberId(memberID).
			Patch(ctx, patch, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
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

	_, err := sender.SendRequest(ctx, call, c.senderCfg)
	return err
}

func (c *channelAPI) ListMessagesNext(ctx context.Context, teamID, channelID, nextLink string, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			WithUrl(nextLink).
			Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}
	if !includeSystem {
		filtered := filterOutSystemEvents(out)
		out.SetValue(filtered)
	}

	return out, nil
}

func (c *channelAPI) ListRepliesNext(ctx context.Context, teamID, channelID, messageID, nextLink string, includeSystem bool) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			ByChatMessageId(messageID).
			Replies().
			WithUrl(nextLink).
			Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, c.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}
	if !includeSystem {
		filtered := filterOutSystemEvents(out)
		out.SetValue(filtered)
	}

	return out, nil
}
