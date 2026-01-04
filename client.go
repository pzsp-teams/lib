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

var graphClient *graph.GraphServiceClient

func getGraphClient(authCfg *config.AuthConfig) (*graph.GraphServiceClient, error) {
	if graphClient == nil {
		tokenProvider, err := auth.GetMSALTokenProvider(authCfg)
		if err != nil {
			return nil, err
		}

		graphClient, err = graph.NewGraphServiceClientWithCredentials(tokenProvider, authCfg.Scopes)
		if err != nil {
			return nil, fmt.Errorf("creating graph client: %w", err)
		}
	}
	return graphClient, nil
}

// NewClient will be used later
func NewClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (*Client, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}

	return NewClientFromGraphClient(cl, senderCfg, cacheCfg)
}

func NewClientFromGraphClient(graphClient *graph.GraphServiceClient, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (*Client, error) {
	teamsAPI := api.GetTeamAPI(graphClient, senderCfg)
	channelAPI := api.GetChannelAPI(graphClient, senderCfg)
	chatAPI := api.GetChatAPI(graphClient, senderCfg)
	userAPI := api.GetUserAPI(graphClient, senderCfg)

	cacheHandler := cacher.GetCacheHandler(cacheCfg)

	// TODO: make resolvers with cache decorators, the same as services
	teamResolver := resolver.GetTeamResolver(teamsAPI, cacheHandler)
	channelResolver := resolver.GetChannelResolver(channelAPI, cacheHandler)
	chatResolver := resolver.GetChatResolver(chatAPI, cacheHandler)

	channelOps := channels.NewOps(channelAPI, userAPI)
	teamOps := teams.NewOps(teamsAPI)
	chatOps := chats.NewOps(chatAPI, userAPI)

	if cacheHandler != nil {
		channelOps = channels.NewOpsWithCache(channelOps, cacheHandler)
		teamOps = teams.NewOpsWithCache(teamOps, cacheHandler)
		chatOps = chats.NewOpsWithCache(chatOps, cacheHandler)
	}
	channelSvc := channels.NewService(channelOps, teamResolver, channelResolver)
	teamSvc := teams.NewService(teamOps, teamResolver)
	chatSvc := chats.NewService(chatOps, chatResolver)

	return &Client{
		Channels: channelSvc,
		Teams:    teamSvc,
		Chats:    chatSvc,
	}, nil
}

func NewChannelServiceFromGraphClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (channels.Service, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}
	channelAPI := api.GetChannelAPI(cl, senderCfg)
	userAPI := api.GetUserAPI(cl, senderCfg)
	teamAPI := api.GetTeamAPI(cl, senderCfg)

	cacheHandler := cacher.GetCacheHandler(cacheCfg)
	teamResolver := resolver.GetTeamResolver(teamAPI, cacheHandler)
	channelResolver := resolver.GetChannelResolver(channelAPI, cacheHandler)

	channelOps := channels.NewOps(channelAPI, userAPI)
	if cacheHandler != nil {
		channelOps = channels.NewOpsWithCache(channelOps, cacheHandler)
	}
	channelSvc := channels.NewService(channelOps, teamResolver, channelResolver)

	return channelSvc, nil
}

func NewTeamServiceFromGraphClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (teams.Service, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}
	teamAPI := api.GetTeamAPI(cl, senderCfg)

	cacheHandler := cacher.GetCacheHandler(cacheCfg)
	teamResolver := resolver.GetTeamResolver(teamAPI, cacheHandler)

	teamOps := teams.NewOps(teamAPI)
	if cacheHandler != nil {
		teamOps = teams.NewOpsWithCache(teamOps, cacheHandler)
	}
	teamSvc := teams.NewService(teamOps, teamResolver)

	return teamSvc, nil
}

func NewChatServiceFromGraphClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (chats.Service, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}
	chatAPI := api.GetChatAPI(cl, senderCfg)
	userAPI := api.GetUserAPI(cl, senderCfg)

	cacheHandler := cacher.GetCacheHandler(cacheCfg)
	chatResolver := resolver.GetChatResolver(chatAPI, cacheHandler)

	chatOps := chats.NewOps(chatAPI, userAPI)
	if cacheHandler != nil {
		chatOps = chats.NewOpsWithCache(chatOps, cacheHandler)
	}
	chatSvc := chats.NewService(chatOps, chatResolver)

	return chatSvc, nil
}

// Close waits for all background operations to complete.
func Close() {
	if cacher.Singleton != nil {
		cacher.Singleton.Runner.Wait()
	}
}
