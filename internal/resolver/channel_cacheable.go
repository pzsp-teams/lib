package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
)

type ChannelResolverCacheable struct {
	cacher   cacher.Cacher
	resolver ChannelResolver
}

func NewChannelResolverCacheable(c cacher.Cacher, r ChannelResolver) ChannelResolver {
	return &ChannelResolverCacheable{
		cacher:   c,
		resolver: r,
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
	key := cacher.NewChannelKeyBuilder(teamID, ref).ToString()
	value, found, err := res.cacher.Get(key)
	if err == nil && found {
		if id, ok := value.(string); ok && id != "" {
			return id, nil
		}
	}
	id, err := res.resolver.ResolveChannelRefToID(ctx, teamID, ref)
	if err != nil {
		return "", err
	}
	_ = res.cacher.Set(key, id)
	return id, nil
}

func (res *ChannelResolverCacheable) ResolveUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return "", fmt.Errorf("empty user reference")
	}
	key := cacher.NewMemberKeyBuilder(ref, teamID, channelID).ToString()
	value, found, err := res.cacher.Get(key)
	if err == nil && found {
		if id, ok := value.(string); ok && id != "" {
			return id, nil
		}
	}
	id, err := res.resolver.ResolveUserRefToMemberID(ctx, teamID, channelID, ref)
	if err != nil {
		return "", err
	}
	_ = res.cacher.Set(key, id)
	return id, nil
}
