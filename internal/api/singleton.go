package api

import (
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/pzsp-teams/lib/config"
)

var (
	channelSingleton ChannelAPI
	chatSingleton    ChatAPI
	teamSingleton    TeamAPI
	userSingleton    UserAPI
	searchSingleton  SearchAPI
)

func GetChannelAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig, searchAPI SearchAPI) ChannelAPI {
	if channelSingleton == nil {
		channelSingleton = NewChannels(c, sCfg, searchAPI)
	}
	return channelSingleton
}

func GetChatAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig, searchAPI SearchAPI) ChatAPI {
	if chatSingleton == nil {
		chatSingleton = NewChat(c, sCfg, searchAPI)
	}
	return chatSingleton
}

func GetTeamAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig) TeamAPI {
	if teamSingleton == nil {
		teamSingleton = NewTeams(c, sCfg)
	}
	return teamSingleton
}

func GetUserAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig) UserAPI {
	if userSingleton == nil {
		userSingleton = NewUser(c, sCfg)
	}
	return userSingleton
}

func GetSearchAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig) SearchAPI {
	if searchSingleton == nil {
		searchSingleton = NewSearch(c, sCfg)
	}
	return searchSingleton
}
