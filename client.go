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
	Chats    chats.Service
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

	return NewClientFromGraphClient(graphClient, senderConfig, cacheEnabled, cachePath)
}

func NewClientFromGraphClient(graphClient *graph.GraphServiceClient, senderConfig *SenderConfig, cacheEnabled bool, cachePath *string) (*Client, error) {
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
	userAPI := api.NewUsers(graphClient, techParams)

	teamResolver := resolver.NewTeamResolverCacheable(teamsAPI, cache, cacheEnabled)
	channelResolver := resolver.NewChannelResolverCacheable(channelsAPI, cache, cacheEnabled)
	chatResolver := resolver.NewChatResolverCacheable(chatAPI, cache, cacheEnabled)

	teamSvc := teams.NewService(teamsAPI, teamResolver)
	channelSvc := channels.NewService(channelsAPI, teamResolver, channelResolver, userAPI)
	chatSvc := chats.NewService(chatAPI, chatResolver, userAPI)
	waitFns := make([]func(), 0, 2)
	if cacheEnabled {
		teamSvc = teams.NewAsyncServiceWithCache(teamSvc, cache)
		channelSvc = channels.NewAsyncServiceWithCache(channelSvc, cache, teamResolver, channelResolver)
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
		waitFns:  waitFns,
	}, nil
}

func (c *Client) Close() {
	for _, fn := range c.waitFns {
		fn()
	}
}
