package resolver

import (
	"context"
	"fmt"
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

func resolveChannelIDByName(chans msmodels.ChannelCollectionResponseable, teamID, ref string) (string, error) {
	if chans == nil || chans.GetValue() == nil || len(chans.GetValue()) == 0 {
		return "", fmt.Errorf("no channels available in team %q", teamID)
	}
	matches := make([]msmodels.Channelable, 0, len(chans.GetValue()))
	for _, c := range chans.GetValue() {
		if c == nil {
			continue
		}
		if util.Deref(c.GetDisplayName()) == ref {
			matches = append(matches, c)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("channel with name %q not found in team %q", ref, teamID)
	case 1:
		id := util.Deref(matches[0].GetId())
		if id == "" {
			return "", fmt.Errorf("channel %q has nil id in team %q", ref, teamID)
		}
		return id, nil
	default:
		var options []string
		for _, c := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", util.Deref(c.GetDisplayName()), util.Deref(c.GetId())))
		}
		return "", fmt.Errorf(
			"multiple channels named %q found in team %q: \n%s.\nPlease use one of the IDs instead",
			ref, teamID, strings.Join(options, ";\n"),
		)
	}
}

func resolveMemberID(members msmodels.ConversationMemberCollectionResponseable, ref string) (string, error) {
	if members == nil || members.GetValue() == nil || len(members.GetValue()) == 0 {
		return "", fmt.Errorf("no members available")
	}
	for _, member := range members.GetValue() {
		if member == nil {
			continue
		}
		um, ok := member.(msmodels.AadUserConversationMemberable)
		if !ok {
			continue
		}
		if matchesUserRef(um, ref) {
			return util.Deref(member.GetId()), nil
		}
	}
	return "", fmt.Errorf("member with reference %q not found", ref)
}

func matchesUserRef(um msmodels.AadUserConversationMemberable, userRef string) bool {
	if userRef == "" {
		return false
	}
	if util.Deref(um.GetUserId()) == userRef {
		return true
	}
	if util.Deref(um.GetDisplayName()) == userRef {
		return true
	}
	email, err := um.GetBackingStore().Get("email")
	if err == nil {
		if emailStr, ok := email.(*string); ok {
			if util.Deref(emailStr) == userRef {
				return true
			}
		}
	}
	return false
}
