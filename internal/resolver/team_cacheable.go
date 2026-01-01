package resolver

import (
	"context"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

// TeamResolver defines methods to resolve team references
// into their corresponding Microsoft Graph IDs.
type TeamResolver interface {
	// ResolveTeamRefToID resolves  a team reference (name or ID) to a team ID.
	//
	// If the reference already appears to be an team ID,
	// it may be returned directly.
	ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error)
	ResolveTeamMemberRefToID(
		ctx context.Context,
		teamID, userRef string,
	) (string, error)
}

// TeamResolverCacheable resolves team references using the graph API
// and optionally caches successful resolutions.
type TeamResolverCacheable struct {
	teamsAPI     api.TeamAPI
	cacheHandler *cacher.CacheHandler
}

// NewTeamResolverCacheable creates a new TeamResolverCacheable.
func NewTeamResolverCacheable(
	teamsAPI api.TeamAPI,
	cacheHandler *cacher.CacheHandler,
) TeamResolver {
	return &TeamResolverCacheable{
		teamsAPI:     teamsAPI,
		cacheHandler: cacheHandler,
	}
}

// ResolveTeamRefToID implements TeamResolver.
func (r *TeamResolverCacheable) ResolveTeamRefToID(
	ctx context.Context,
	teamRef string,
) (string, error) {
	rCtx := r.newTeamResolveContext(teamRef)
	return rCtx.resolveWithCache(ctx, r.cacheHandler)
}

func (r *TeamResolverCacheable) newTeamResolveContext(
	teamRef string,
) resolverContext[msmodels.TeamCollectionResponseable] {
	ref := strings.TrimSpace(teamRef)
	return resolverContext[msmodels.TeamCollectionResponseable]{
		cacheKey:    cacher.NewTeamKey(ref),
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

// ResolveTeamMemberRefToID implements TeamResolver.
func (r *TeamResolverCacheable) ResolveTeamMemberRefToID(
	ctx context.Context,
	teamID, userRef string,
) (string, error) {
	rCtx := r.newTeamMemberResolveContext(teamID, userRef)
	return rCtx.resolveWithCache(ctx, r.cacheHandler)
}

func (r *TeamResolverCacheable) newTeamMemberResolveContext(
	teamID, userRef string,
) resolverContext[msmodels.ConversationMemberCollectionResponseable] {
	ref := strings.TrimSpace(userRef)
	return resolverContext[msmodels.ConversationMemberCollectionResponseable]{
		cacheKey:    cacher.NewTeamMemberKey(teamID, ref, nil),
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyGUID(ref) },
		fetch: func(ctx context.Context) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
			return r.teamsAPI.ListMembers(ctx, teamID)
		},
		extract: func(data msmodels.ConversationMemberCollectionResponseable) (string, error) {
			return resolveMemberID(data, ref)
		},
	}
}