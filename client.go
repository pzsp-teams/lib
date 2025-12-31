package lib

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/auth"
	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/teams"
)

type Client struct {
	Channels channels.Service
	Teams    teams.Service
	Chats    chats.Service
}

// NewClient will be used later
func NewClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (*Client, error) {
	tokenProvider, err := auth.NewMSALTokenProvider(authCfg)
	if err != nil {
		return nil, fmt.Errorf("creating token provider: %w", err)
	}

	graphClient, err := graph.NewGraphServiceClientWithCredentials(tokenProvider, authCfg.Scopes)
	if err != nil {
		return nil, fmt.Errorf("creating graph client: %w", err)
	}

	teamsAPI := api.NewTeams(graphClient, senderCfg)
	channelsAPI := api.NewChannels(graphClient, senderCfg)
	chatAPI := api.NewChat(graphClient, senderCfg)
	userAPI := api.NewUsers(graphClient, senderCfg)

	cacheHandler := cacher.NewCacheHandler(cacheCfg)

	// TODO: make resolvers with cache decorators, the same as services
	teamResolver := resolver.NewTeamResolverCacheable(teamsAPI, cacheHandler)
	channelResolver := resolver.NewChannelResolverCacheable(channelsAPI, cacheHandler)
	chatResolver := resolver.NewChatResolverCacheable(chatAPI, cacheHandler)

	teamSvc := teams.NewService(teamsAPI, teamResolver)
	channelSvc := channels.NewService(channelsAPI, teamResolver, channelResolver, userAPI)
	chatSvc := chats.NewService(chatAPI, chatResolver, userAPI)

	if cacheHandler != nil {
		channelSvc = channels.NewServiceWithCache(channelSvc, teamResolver, channelResolver, cacheHandler)
		teamSvc = teams.NewServiceWithCache(teamSvc, cacheHandler)
	}

	return &Client{
		Channels: channelSvc,
		Teams:    teamSvc,
		Chats:    chatSvc,
	}, nil
}

func (c *Client) Close() {
	if w, ok := c.Channels.(interface{ Wait() }); ok {
		w.Wait()
	}

	if w, ok := c.Teams.(interface{ Wait() }); ok {
		w.Wait()
	}
}
