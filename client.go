package teams

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/auth"
	"github.com/pzsp-teams/lib/internal/mapper"

	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/teams"
)

type Client struct {
	Channels *channels.Service
	Teams    *teams.Service
}

// NewClient will be used later
func NewClient(ctx context.Context, authConfig *AuthConfig, senderConfig *SenderConfig) (*Client, error) {
	tokenProvider, err := auth.NewMSALTokenProvider(authConfig.toMSALCredentials())
	if err != nil {
		return nil, fmt.Errorf("creating token provider: %w", err)
	}

	graphClient, err := graph.NewGraphServiceClientWithCredentials(tokenProvider, authConfig.Scopes)
	if err != nil {
		return nil, fmt.Errorf("creating graph client: %w", err)
	}

	techParams := senderConfig.toTechParams()

	teamsAPI := api.NewTeams(graphClient, techParams)
	channelsAPI := api.NewChannels(graphClient, techParams)

	teamMapper := mapper.NewTeamMapper(teamsAPI, channelsAPI)
	channelMapper := mapper.NewChannelMapper(channelsAPI)

	teamSvc := teams.NewService(teamsAPI, teamMapper)
	chSvc := channels.NewService(channelsAPI, teamMapper, channelMapper)

	return &Client{Channels: chSvc, Teams: teamSvc}, nil
}
