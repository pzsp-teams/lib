package teams

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/pzsp-teams/lib/internal/auth"
	"github.com/pzsp-teams/lib/pkg/teams/channels"
	"github.com/pzsp-teams/lib/pkg/teams/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

// Client will be used later
type Client struct {
	Channels *channels.Service
	Teams *teams.Service
}

// SenderConfig will be used later
type SenderConfig struct {
	MaxRetries     int
	NextRetryDelay int
	Timeout        int
}

// NewClient will be used later
func NewClient(ctx context.Context, authConfig *AuthConfig, senderConfig *SenderConfig) (*Client, error) {
	tokenProvider, err := auth.NewMSALTokenProvider(&auth.MSALCredentials{
		ClientID: authConfig.ClientID,
		Tenant:   authConfig.Tenant,
		Scopes:   authConfig.Scopes,
	})
	if err != nil {
		return nil, err
	}
	cred := &msalCredential{
		provider: tokenProvider,
		email:    authConfig.Email,
	}

	graphClient, err := graph.NewGraphServiceClientWithCredentials(cred, authConfig.Scopes)
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
	channelAPI := channels.NewChannelsAPI(graphClient, techParams)
	teamsAPI := teams.NewTeamsAPI(graphClient, techParams)
	chSvc := channels.NewService(channelAPI)
	teamSvc := teams.NewService(teamsAPI)

	return &Client{
		Channels: chSvc,
		Teams:    teamSvc,
	}
}
