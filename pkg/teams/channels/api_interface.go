package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

// ChannelAPIInterface will be used later
type ChannelAPIInterface interface {
	ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError)
	GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError)
	CreateChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError)
	DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError
}
