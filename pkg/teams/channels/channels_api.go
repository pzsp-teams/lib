package channels

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	models "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

// API will be used later
type API struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

// NewChannelsAPI will be used later
func NewChannelsAPI(client *graph.GraphServiceClient, techParams sender.RequestTechParams) *API {
	return &API{client, techParams}
}

// ListChannels will be used later
func (api *API) ListChannels(ctx context.Context, teamID string) (models.ChannelCollectionResponseable, *sender.RequestError) {
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

	out, ok := resp.(models.ChannelCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChannelCollectionResponseable"}
	}

	return out, nil
}

// GetChannel will be used later
func (api *API) GetChannel(ctx context.Context, teamID, channelID string) (models.Channelable, *sender.RequestError) {
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

	out, ok := resp.(models.Channelable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Channelable"}
	}

	return out, nil
}

// CreateChannel will be used later
func (api *API) CreateChannel(ctx context.Context, teamID string, channel models.Channelable) (models.Channelable, *sender.RequestError) {
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

	out, ok := resp.(models.Channelable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Channelable"}
	}

	return out, nil
}

// DeleteChannel will be used later
func (api *API) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
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
func (api *API) SendMessage(ctx context.Context, teamID, channelID string, message models.ChatMessageable) (models.ChatMessageable, *sender.RequestError) {
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

	out, ok := resp.(models.ChatMessageable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageable"}
	}

	return out, nil
}

// ListMessages will be used later
func (api *API) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (models.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		queryParameters := &teams.ItemChannelsItemMessagesRequestBuilderGetQueryParameters{}
		if top != nil {
			queryParameters.Top = top
		}
		return api.client.
			Teams().
			ByTeamId(teamID).
			Channels().
			ByChannelId(channelID).
			Messages().
			Get(ctx, &teams.ItemChannelsItemMessagesRequestBuilderGetRequestConfiguration{
				QueryParameters: queryParameters,
			})
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(models.ChatMessageCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageCollectionResponseable"}
	}

	return out, nil
}

// GetMessage will be used later
func (api *API) GetMessage(ctx context.Context, teamID, channelID, messageID string) (models.ChatMessageable, *sender.RequestError) {
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

	out, ok := resp.(models.ChatMessageable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageable"}
	}

	return out, nil
}

// ListReplies will be used later
func (api *API) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (models.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		queryParameters := &teams.ItemChannelsItemMessagesItemRepliesRequestBuilderGetQueryParameters{}
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
			Get(ctx, &teams.ItemChannelsItemMessagesItemRepliesRequestBuilderGetRequestConfiguration{
				QueryParameters: queryParameters,
			})
	}

	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(models.ChatMessageCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageCollectionResponseable"}
	}

	return out, nil
}

// GetReply will be used later
func (api *API) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (models.ChatMessageable, *sender.RequestError) {
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

	out, ok := resp.(models.ChatMessageable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatMessageable"}
	}

	return out, nil
}
