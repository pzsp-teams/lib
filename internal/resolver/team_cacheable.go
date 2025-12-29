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

func resolveTeamIDByName(list msmodels.TeamCollectionResponseable, ref string) (string, error) {
	if list == nil || list.GetValue() == nil || len(list.GetValue()) == 0 {
		return "", fmt.Errorf("no teams available")
	}
	matches := make([]msmodels.Teamable, 0, len(list.GetValue()))
	for _, t := range list.GetValue() {
		if t == nil {
			continue
		}
		if util.Deref(t.GetDisplayName()) == ref {
			matches = append(matches, t)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("team with name %q not found", ref)
	case 1:
		id := util.Deref(matches[0].GetId())
		if id == "" {
			return "", fmt.Errorf("team %q has nil id", ref)
		}
		return id, nil
	default:
		var options []string
		for _, t := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", util.Deref(t.GetDisplayName()), util.Deref(t.GetId())))
		}
		return "", fmt.Errorf(
			"multiple teams named %q found: \n%s.\nPlease use one of the IDs instead",
			ref, strings.Join(options, ";\n"),
		)
	}
}
