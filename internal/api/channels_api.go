package api

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	models "github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/pzsp-teams/lib/internal/sender"
)

type ChannelsAPI struct {
    client     *graph.GraphServiceClient
    techParams sender.RequestTechParams
}

func NewChannelsAPI(client *graph.GraphServiceClient, techParams sender.RequestTechParams) *ChannelsAPI {
    return &ChannelsAPI{client, techParams}
}

func (api *ChannelsAPI) ListChannels(ctx context.Context, teamId string) (models.ChannelCollectionResponseable, *sender.RequestError) {
    call := func(ctx context.Context) (sender.Response, error) {
        return api.client.
            Teams().
            ByTeamId(teamId).
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

func (api *ChannelsAPI) GetChannel(ctx context.Context, teamId, channelId string) (models.Channelable, *sender.RequestError) {
    call := func(ctx context.Context) (sender.Response, error) {
        return api.client.
            Teams().
            ByTeamId(teamId).
            Channels().
            ByChannelId(channelId).
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

func (api *ChannelsAPI) CreateChannel(ctx context.Context, teamId string, channel models.Channelable) (models.Channelable, *sender.RequestError) {
    call := func(ctx context.Context) (sender.Response, error) {
        return api.client.
            Teams().
            ByTeamId(teamId).
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

func (api *ChannelsAPI) DeleteChannel(ctx context.Context, teamId, channelId string) *sender.RequestError {
    call := func(ctx context.Context) (sender.Response, error) {
        err := api.client.
            Teams().
            ByTeamId(teamId).
            Channels().
            ByChannelId(channelId).
            Delete(ctx, nil)
        return nil, err
    }

    _, err := sender.SendRequest(ctx, call, api.techParams)
    return err
}
