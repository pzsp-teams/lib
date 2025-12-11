package lib

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/auth"
	"github.com/pzsp-teams/lib/internal/resolver"

	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/teams"
)

type Client struct {
	Channels *channels.Service
	Teams    *teams.Service
	Chats    *chats.Service
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

	return NewClientFromGraphClient(graphClient, senderConfig)
}

// NewClientFromGraphClient will be used later
func NewClientFromGraphClient(graphClient *graph.GraphServiceClient, senderConfig *SenderConfig) (*Client, error) {
	techParams := senderConfig.toTechParams()

	teamsAPI := api.NewTeams(graphClient, techParams)
	channelsAPI := api.NewChannels(graphClient, techParams)
	chatAPI := api.NewChat(graphClient, techParams)

	teamMapper := resolver.NewTeamResolver(teamsAPI, channelsAPI)
	channelMapper := resolver.NewChannelResolver(channelsAPI)

	teamSvc := teams.NewService(teamsAPI, teamMapper)
	channelSvc := channels.NewService(channelsAPI, teamMapper, channelMapper)
	chatSvc := chats.NewService(chatAPI)

	return &Client{
		Channels: channelSvc,
		Teams:    teamSvc,
		Chats:    chatSvc,
	}, nil
}
