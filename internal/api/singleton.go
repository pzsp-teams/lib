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
)

func GetChannelAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig) ChannelAPI {
	if channelSingleton == nil {
		channelSingleton = NewChannels(c, sCfg)
	}
	return channelSingleton
}

func GetChatAPI(c *graph.GraphServiceClient, sCfg *config.SenderConfig) ChatAPI {
	if chatSingleton == nil {
		chatSingleton = NewChat(c, sCfg)
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
