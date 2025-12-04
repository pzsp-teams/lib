package teams

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/pzsp-teams/lib/internal/auth"
	channelsAPI "github.com/pzsp-teams/lib/internal/channels"
	"github.com/pzsp-teams/lib/internal/mapper"
	teamsAPI "github.com/pzsp-teams/lib/internal/teams"
	"github.com/pzsp-teams/lib/pkg/teams/channels"
	"github.com/pzsp-teams/lib/pkg/teams/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

// Client will be used later
type Client struct {
	Channels *channels.Service
	Teams    *teams.Service
}

// SenderConfig will be used later
type SenderConfig struct {
	MaxRetries     int
	NextRetryDelay int
	Timeout        int
}

// NewClient will be used later
func NewClient(ctx context.Context, authConfig *auth.AuthConfig, senderConfig *SenderConfig) (*Client, error) {
	tokenProvider, err := auth.NewMSALTokenProvider(authConfig)
	if err != nil {
		return nil, err
	}

	graphClient, err := graph.NewGraphServiceClientWithCredentials(tokenProvider, authConfig.Scopes)
	if err != nil {
		return nil, err
	}

	return newClient(graphClient, senderConfig), nil
}

func newClient(graphClient *graph.GraphServiceClient, opts *SenderConfig) *Client {
	techParams := sender.RequestTechParams{
		MaxRetries:     opts.MaxRetries,
		NextRetryDelay: opts.NextRetryDelay,
		Timeout:        opts.Timeout,
	}
	channelAPI := channelsAPI.NewChannelsAPI(graphClient, techParams)
	teamAPI := teamsAPI.NewTeamsAPI(graphClient, techParams)
	nameMapper := mapper.NewMapper(teamAPI, channelAPI)
	chSvc := channels.NewService(channelAPI, nameMapper)
	teamSvc := teams.NewService(teamAPI, nameMapper)

	return &Client{
		Channels: chSvc,
		Teams:    teamSvc,
	}
}
