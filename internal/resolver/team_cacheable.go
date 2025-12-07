package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
)

type TeamResolverCacheable struct {
	cacher cacher.Cacher
	resolver TeamResolver
}

func NewTeamResolverCacheable(cacher cacher.Cacher, resolver TeamResolver) TeamResolver {
	return &TeamResolverCacheable{
		cacher: cacher,
		resolver: resolver,
	}
}

func (res *TeamResolverCacheable) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	ref := strings.TrimSpace(teamRef)
	if ref == "" {
		return "", fmt.Errorf("empty team reference")
	}
	if isLikelyGUID(ref) {
		return ref, nil
	}
	keyBuilder := cacher.NewTeamKeyBuilder(ref)
	key := keyBuilder.ToString()
	value, found, err := res.cacher.Get(key)
	if err == nil && found {
		if id, ok := value.(string); ok && id != "" {
			return id, nil
		}
	}
	id, err := res.resolver.ResolveTeamRefToID(ctx, ref)
	if err != nil {
		return "", err
	}
	_ = res.cacher.Set(key, id)
	return id, nil
}