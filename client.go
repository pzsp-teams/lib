// Package lib acts as the primary entry point for the Microsoft Teams API client library.
// It adopts a Facade pattern, aggregating specialized services (Teams, Channels, Chats)
// into a single, cohesive Client.
//
// The package manages the complexity of:
//   - Authentication (via MSAL and Graph Token Providers).
//   - Dependency Injection (wiring APIs, Caches, and Resolvers).
//   - Caching strategies (transparently wrapping operations with caching layers).
//
// Usage:
// Initialize the Client using NewClient for a standard setup.
// Alternatively, if you need only specific services, use:
//   - NewTeamServiceFromGraphClient for Teams service.
//   - NewChannelServiceFromGraphClient for Channels service.
//   - NewChatServiceFromGraphClient for Chats service.
//
// Always ensure to call Close() upon application shutdown to flush any background
// cache operations.
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

// Client is the central hub for interacting with the Microsoft Teams ecosystem.
// It aggregates access to specific domains: Channels, Teams, and Chats, hiding
// the complexity of underlying Graph API calls and caching mechanisms.
type Client struct {
	Channels channels.Service
	Teams    teams.Service
	Chats    chats.Service
}

// graphClient is a package-level singleton to hold the authenticated Graph client.
// Note: This approach assumes a single identity per application instance.
var graphClient *graph.GraphServiceClient

// getGraphClient ensures a singleton instance of the GraphServiceClient is created.
// It initializes the MSAL token provider using the provided authentication config.
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

// NewClient initializes a new Client instance with fully configured internal services.
// It handles the authentication handshake using the provided authCfg and sets up
// sending and caching behaviors based on senderCfg and cacheCfg.
func NewClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (*Client, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}

	return NewClientFromGraphClient(cl, senderCfg, cacheCfg)
}

// NewClientFromGraphClient creates a Client using an existing, pre-configured GraphServiceClient.
// This is a separated exported constructor mainly for external testing purposes (via mocking Teams API by injection of GraphServiceClient).
//
// It wires up all internal dependencies, including API clients, caching layers, and
// entity resolvers (e.g., resolving team names to IDs).
func NewClientFromGraphClient(graphClient *graph.GraphServiceClient, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (*Client, error) {
	teamsAPI := api.GetTeamAPI(graphClient, senderCfg)
	searchAPI := api.GetSearchAPI(graphClient, senderCfg)
	channelAPI := api.GetChannelAPI(graphClient, senderCfg, searchAPI)
	chatAPI := api.GetChatAPI(graphClient, senderCfg, searchAPI)
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

// NewChannelServiceFromGraphClient creates a standalone service for Channel operations.
// Use this if you do not need the full Client wrapper and only want to interact with Channels.
func NewChannelServiceFromGraphClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (channels.Service, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}
	searchAPI := api.GetSearchAPI(cl, senderCfg)
	channelAPI := api.GetChannelAPI(cl, senderCfg, searchAPI)
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

// NewTeamServiceFromGraphClient creates a standalone service for Team operations.
// Use this if you do not need the full Client wrapper and only want to interact with Teams.
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

// NewChatServiceFromGraphClient creates a standalone service for Chat operations.
// Use this if you do not need the full Client wrapper and only want to interact with Chats.
func NewChatServiceFromGraphClient(ctx context.Context, authCfg *config.AuthConfig, senderCfg *config.SenderConfig, cacheCfg *config.CacheConfig) (chats.Service, error) {
	cl, err := getGraphClient(authCfg)
	if err != nil {
		return nil, err
	}
	searchAPI := api.GetSearchAPI(cl, senderCfg)
	chatAPI := api.GetChatAPI(cl, senderCfg, searchAPI)
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

// Close ensures a graceful shutdown of the library.
// It waits for any pending background operations (such as asynchronous cache updates)
// to complete before returning, preventing data loss or race conditions.
func Close() {
	if cacher.Singleton != nil {
		cacher.Singleton.Runner.Wait()
	}
}
