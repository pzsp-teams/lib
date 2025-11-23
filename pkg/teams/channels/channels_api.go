package channels

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	models "github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/pzsp-teams/lib/internal/sender"
)

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
func (api *ChannelsAPI) ListChannels(ctx context.Context, teamID string) (models.ChannelCollectionResponseable, *sender.RequestError) {
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
func (api *ChannelsAPI) GetChannel(ctx context.Context, teamID, channelID string) (models.Channelable, *sender.RequestError) {
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
func (api *ChannelsAPI) CreateChannel(ctx context.Context, teamID string, channel models.Channelable) (models.Channelable, *sender.RequestError) {
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
