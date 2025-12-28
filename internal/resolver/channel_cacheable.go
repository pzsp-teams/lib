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
}
type ChannelResolverCacheable struct {
	cacher       cacher.Cacher
	channelsAPI  api.ChannelAPI
	cacheEnabled bool
}

func NewChannelResolverCacheable(a api.ChannelAPI, c cacher.Cacher, cacheEnabled bool) ChannelResolver {
	return &ChannelResolverCacheable{
		cacher:       c,
		channelsAPI:  a,
		cacheEnabled: cacheEnabled,
	}
}

func (res *ChannelResolverCacheable) ResolveChannelRefToID(ctx context.Context, teamID, channelRef string) (string, error) {
	ref := strings.TrimSpace(channelRef)
	if ref == "" {
		return "", fmt.Errorf("empty channel reference")
	}
	if util.IsLikelyChannelID(ref) {
		return ref, nil
	}
	if res.cacheEnabled {
		key := cacher.NewChannelKey(teamID, ref)
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
		key := cacher.NewChannelKey(teamID, ref)
		_ = res.cacher.Set(key, idResolved)
	}
	return idResolved, nil
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
