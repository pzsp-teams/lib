package api

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

// ChannelsAPIInterface will be used later
type Channels interface {
	ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError)
	GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError)
	CreateChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError)
	DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError
	SendMessage(ctx context.Context, teamID, channelID string, message msmodels.ChatMessageable) (msmodels.ChatMessageable, *sender.RequestError)
	ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError)
	ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError)
}

// ChannelsAPI will be used later
type ChannelsAPI struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

// NewChannelsAPI will be used later
func NewChannelsAPI(client *graph.GraphServiceClient, techParams sender.RequestTechParams) *ChannelsAPI {
	return &ChannelsAPI{client, techParams}
}

// ListChannels will be used later
func (api *ChannelsAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChannelCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChannelCollectionResponseable"}
	}

	return out, nil
}

// GetChannel will be used later
func (api *ChannelsAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Channelable"}
	}

	return out, nil
}

// CreateChannel will be used later
func (api *ChannelsAPI) CreateChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			Post(ctx, channel, nil)
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Channelable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Channelable"}
	}

	return out, nil
}

// DeleteChannel will be used later
func (api *ChannelsAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		err := api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Delete(ctx, nil)
		return nil, err
	}

	_, err := sender.SendRequest(ctx, call, api.techParams)
	return err
}

// SendMessage will be used later
func (api *ChannelsAPI) SendMessage(ctx context.Context, teamID, channelID string, message msmodels.ChatMessageable) (msmodels.ChatMessageable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			Post(ctx, message, nil)
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageable"}
	}

	return out, nil
}

// ListMessages will be used later
func (api *ChannelsAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		queryParameters := &graphteams.ItemChannelsItemMessagesRequestBuilderGetQueryParameters{}
		if top != nil {
			queryParameters.Top = top
		}
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			Get(ctx, &graphteams.ItemChannelsItemMessagesRequestBuilderGetRequestConfiguration{
				QueryParameters: queryParameters,
			})
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageCollectionResponseable"}
	}

	return out, nil
}

// GetMessage will be used later
func (api *ChannelsAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			ByChatMessageId(messageID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageable"}
	}

	return out, nil
}

// ListReplies will be used later
func (api *ChannelsAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		queryParameters := &graphteams.ItemChannelsItemMessagesItemRepliesRequestBuilderGetQueryParameters{}
		if top != nil {
			queryParameters.Top = top
		}
		return api.client.
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

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageCollectionResponseable"}
	}

	return out, nil
}

// GetReply will be used later
func (api *ChannelsAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
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

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageable"}
	}

	return out, nil
}
