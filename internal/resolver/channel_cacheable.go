package resolver

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/util"
)

type ChannelResolver interface {
	ResolveChannelRefToID(ctx context.Context, teamID, channelName string) (string, error)
	ResolveUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error)
}
type ChannelResolverCacheable struct {
	cacher      cacher.Cacher
	channelsAPI api.ChannelAPI
	cacheEnabled bool
}

func NewChannelResolverCacheable(a api.ChannelAPI, c cacher.Cacher, cacheEnabled bool) ChannelResolver {
	return &ChannelResolverCacheable{
		cacher:      c,
		channelsAPI: a,
		cacheEnabled: cacheEnabled,
	}
}

func (res *ChannelResolverCacheable) ResolveChannelRefToID(ctx context.Context, teamID, channelRef string) (string, error) {
	ref := strings.TrimSpace(channelRef)
	if ref == "" {
		return "", fmt.Errorf("empty channel reference")
	}
	if isLikelyChannelID(ref) {
		return ref, nil
	}
	if res.cacheEnabled {
		key := cacher.NewChannelKeyBuilder(teamID, ref).ToString()
		value, found, cacheErr := res.cacher.Get(key)
		if cacheErr == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 {
				return ids[0], nil
			}
		}
	}
	chans, err := res.channelsAPI.ListChannels(ctx, teamID)
	if err != nil {
		return "", err
	}
	idResolved, senderErr := resolveChannelIDByName(teamID, ref, chans)
	if senderErr != nil {
		return "", senderErr
	}
	if res.cacheEnabled {
		key := cacher.NewChannelKeyBuilder(teamID, ref).ToString()
		_ = res.cacher.Set(key, idResolved)
	}
	return idResolved, nil
}

func (res *ChannelResolverCacheable) ResolveUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return "", fmt.Errorf("empty user reference")
	}
	if res.cacheEnabled {
		key := cacher.NewMemberKeyBuilder(ref, teamID, channelID).ToString()
		value, found, err := res.cacher.Get(key)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 {
				return ids[0], nil
			}
		}
	}
	resp, apiErr := res.channelsAPI.ListMembers(ctx, teamID, channelID)
	if apiErr != nil {
		return "", apiErr
	}
	if resp == nil || resp.GetValue() == nil || len(resp.GetValue()) == 0 {
		return "", fmt.Errorf("no members found in channel %q", channelID)
	}
	id := ""
	for _, member := range resp.GetValue() {
		if member == nil {
			continue
		}
		um, ok := member.(msmodels.AadUserConversationMemberable)
		if !ok {
			continue
		}
		if matchesUserRef(um, ref) {
			id = util.Deref(member.GetId())
			break
		}
	}
	if id == "" {
		return "", fmt.Errorf("user %q not found in channel %q", ref, channelID)
	}
	if res.cacheEnabled {
		key := cacher.NewMemberKeyBuilder(ref, teamID, channelID).ToString()
		_ = res.cacher.Set(key, id)
	}
	return id, nil
}

func isLikelyChannelID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@thread.")
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
	ad := um.GetAdditionalData()
	if ad == nil {
		return false
	}
	for _, key := range []string{"userPrincipalName", "mail", "email"} {
		raw, ok := ad[key]
		if !ok {
			continue
		}
		if v, ok := raw.(string); ok && strings.EqualFold(v, userRef) {
			return true
		}
	}
	return false
}

func resolveChannelIDByName(teamID, ref string, chans msmodels.ChannelCollectionResponseable) (string, error) {
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
