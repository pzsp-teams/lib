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

type TeamResolver interface {
	ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error)
}

type TeamResolverCacheable struct {
	teamsAPI     api.TeamAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

func NewTeamResolverCacheable(teamsAPI api.TeamAPI, cacher cacher.Cacher, cacheEnabled bool) TeamResolver {
	return &TeamResolverCacheable{
		teamsAPI:     teamsAPI,
		cacher:       cacher,
		cacheEnabled: cacheEnabled,
	}
}

func (r *TeamResolverCacheable) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	rCtx := r.newTeamResolveContext(teamRef)
	return rCtx.resolveWithCache(ctx, r.cacher, r.cacheEnabled)
}

func (r *TeamResolverCacheable) newTeamResolveContext(teamRef string) ResolverContext[msmodels.TeamCollectionResponseable] {
	ref := strings.TrimSpace(teamRef)
	return ResolverContext[msmodels.TeamCollectionResponseable]{
		cacheKey:    cacher.NewTeamKey(ref),
		keyType:     cacher.Team,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyGUID(ref) },
		fetch: func(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
			return r.teamsAPI.ListMyJoined(ctx)
		},
		extract: func(data msmodels.TeamCollectionResponseable) (string, error) {
			return resolveTeamIDByName(data, ref)
		},
	}
}
