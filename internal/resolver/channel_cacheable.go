package resolver

import (
	"context"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

type ChannelResolver interface {
	ResolveChannelRefToID(ctx context.Context, teamID, channelName string) (string, error)
	ResolveChannelMemberRefToID(ctx context.Context, teamID, channelID, userRef string) (string, error)
}

type ChannelResolverCacheable struct {
	channelsAPI  api.ChannelAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

func NewChannelResolverCacheable(channelAPI api.ChannelAPI, cacher cacher.Cacher, cacheEnabled bool) ChannelResolver {
	return &ChannelResolverCacheable{
		channelsAPI:  channelAPI,
		cacher:       cacher,
		cacheEnabled: cacheEnabled,
	}
}

func (res *ChannelResolverCacheable) ResolveChannelRefToID(ctx context.Context, teamID, channelRef string) (string, error) {
	rCtx := res.newChannelResolveContext(teamID, channelRef)
	return rCtx.resolveWithCache(ctx, res.cacher, res.cacheEnabled)
}

func (res *ChannelResolverCacheable) ResolveChannelMemberRefToID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	rCtx := res.newChannelMemberResolveContext(teamID, channelID, userRef)
	return rCtx.resolveWithCache(ctx, res.cacher, res.cacheEnabled)
}

func (res *ChannelResolverCacheable) newChannelResolveContext(teamID, channelRef string) ResolverContext[msmodels.ChannelCollectionResponseable] {
	ref := strings.TrimSpace(channelRef)
	return ResolverContext[msmodels.ChannelCollectionResponseable]{
		cacheKey:    cacher.NewChannelKey(teamID, ref),
		keyType:     cacher.Channel,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyThreadConversationID(ref) },
		fetch: func(ctx context.Context) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
			return res.channelsAPI.ListChannels(ctx, teamID)
		},
		extract: func(data msmodels.ChannelCollectionResponseable) (string, error) {
			return resolveChannelIDByName(data, teamID, ref)
		},
	}
}

func (res *ChannelResolverCacheable) newChannelMemberResolveContext(teamID, channelID, userRef string) ResolverContext[msmodels.ConversationMemberCollectionResponseable] {
	ref := strings.TrimSpace(userRef)
	return ResolverContext[msmodels.ConversationMemberCollectionResponseable]{
		cacheKey:    cacher.NewChannelMemberKey(teamID, channelID, ref, nil),
		keyType:     cacher.ChannelMember,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyGUID(ref) },
		fetch: func(ctx context.Context) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
			return res.channelsAPI.ListMembers(ctx, teamID, channelID)
		},
		extract: func(data msmodels.ConversationMemberCollectionResponseable) (string, error) {
			return resolveMemberID(data, ref)
		},
	}
}
