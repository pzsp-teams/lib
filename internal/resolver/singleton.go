package resolver

import (
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/cacher"
)

var (
	chatSingleton ChatResolver
	teamSingleton TeamResolver
	channelSingleton ChannelResolver
)

func GetChatResolver(chatAPI api.ChatAPI, cacheHandler *cacher.CacheHandler) ChatResolver {
	if chatSingleton == nil {
		chatSingleton = NewChatResolverCacheable(chatAPI, cacheHandler)
	}
	return chatSingleton
}

func GetTeamResolver(teamAPI api.TeamAPI, cacheHandler *cacher.CacheHandler) TeamResolver {
	if teamSingleton == nil {
		teamSingleton = NewTeamResolverCacheable(teamAPI, cacheHandler)
	}
	return teamSingleton
}

func GetChannelResolver(channelAPI api.ChannelAPI, cacheHandler *cacher.CacheHandler) ChannelResolver {
	if channelSingleton == nil {
		channelSingleton = NewChannelResolverCacheable(channelAPI, cacheHandler)
	}
	return channelSingleton
}