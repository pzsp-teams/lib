package resolver

import (
	"context"
	"strings"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

type UserResolver interface {
	ResolveUserRefToID(ctx context.Context, userRef string) (string, error)
}


type userResolverCacheable struct {
	userAPI api.UsersAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

func NewUserResolverCacheable(userAPI api.UsersAPI, cacher cacher.Cacher, cacheEnabled bool) UserResolver {
	return &userResolverCacheable{
		userAPI:      userAPI,
		cacher:       cacher,
		cacheEnabled: cacheEnabled,
	}
}

func (r *userResolverCacheable) ResolveUserRefToID(ctx context.Context, userRef string) (string, error) {
	rCtx := r.newUserResolveContext(userRef)
	return rCtx.resolveWithCache(ctx, r.cacher, r.cacheEnabled)
}

func (r *userResolverCacheable) newUserResolveContext(userRef string) ResolverContext[string] {
	ref := strings.TrimSpace(userRef)
	ref = strings.ToLower(ref)
	return ResolverContext[string]{
		cacheKey:    cacher.NewUserKey(ref, nil),
		keyType:     cacher.User,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyGUID(ref) },
		fetch: func(ctx context.Context) (string, *sender.RequestError) {
			return r.userAPI.GetUserIDByEmailOrUPN(ctx, ref)
		},
		extract: func(data string) (string, error) {
			return data, nil
		},
	}
}



