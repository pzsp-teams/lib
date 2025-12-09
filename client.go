package lib

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/auth"
	"github.com/pzsp-teams/lib/internal/resolver"

	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/teams"
)

type Client struct {
	Channels channels.Service
	Teams    teams.Service
	Chats    *chats.Service
	waitFns  []func()
}

// NewClient will be used later
func NewClient(ctx context.Context, authConfig *AuthConfig, senderConfig *SenderConfig, cacheEnabled bool, cachePath *string) (*Client, error) {
	tokenProvider, err := auth.NewMSALTokenProvider(authConfig.toMSALCredentials())
	if err != nil {
		return nil, fmt.Errorf("creating token provider: %w", err)
	}

	graphClient, err := graph.NewGraphServiceClientWithCredentials(tokenProvider, authConfig.Scopes)
	if err != nil {
		return nil, fmt.Errorf("creating graph client: %w", err)
	}
	var cache cacher.Cacher
	if cacheEnabled && cachePath != nil {
		cache = cacher.NewJSONFileCacher(*cachePath)
	} else if cacheEnabled {
		cache = cacher.NewJSONFileCacher(defaultCachePath())
	}
	techParams := senderConfig.toTechParams()

	teamsAPI := api.NewTeams(graphClient, techParams)
	channelsAPI := api.NewChannels(graphClient, techParams)
	chatAPI := api.NewChat(graphClient, techParams)

	teamResolver := resolver.NewTeamResolverCacheable(teamsAPI, cache, cacheEnabled)
	channelResolver := resolver.NewChannelResolverCacheable(channelsAPI, cache, cacheEnabled)

	teamSvc := teams.NewService(teamsAPI, teamResolver)
	channelSvc := channels.NewService(channelsAPI, teamResolver, channelResolver)
	chatSvc := chats.NewService(chatAPI)
	waitFns := make([]func(), 0, 2)
	if cacheEnabled {
		teamSvc = teams.NewAsyncServiceWithAutoCacheManagement(teamSvc, cache)
		channelSvc = channels.NewAsyncServiceWithAutoCacheManagement(channelSvc, cache, teamResolver, channelResolver)
		if w, ok := channelSvc.(interface{ Wait() }); ok {
			waitFns = append(waitFns, w.Wait)
		}
		if w, ok := teamSvc.(interface{ Wait() }); ok {
			waitFns = append(waitFns, w.Wait)
		}
	}
	return &Client{
		Channels: channelSvc,
		Teams:    teamSvc,
		Chats:    chatSvc,
		waitFns: waitFns,
	}, nil
}

func (c *Client) Close() {
	for _, fn := range c.waitFns {
		fn()
	}
}
