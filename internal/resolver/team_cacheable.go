package resolver

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
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

func NewTeamResolverCacheable(teamsAPI api.TeamAPI, c cacher.Cacher, cacheEnabled bool) TeamResolver {
	return &TeamResolverCacheable{
		teamsAPI:     teamsAPI,
		cacher:       c,
		cacheEnabled: cacheEnabled,
	}
}

func (m *TeamResolverCacheable) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	ref := strings.TrimSpace(teamRef)
	if ref == "" {
		return "", fmt.Errorf("empty team reference")
	}
	if isLikelyGUID(ref) {
		return ref, nil
	}
	if m.cacheEnabled && m.cacher != nil {
		key := cacher.NewTeamKeyBuilder(ref).ToString()
		value, found, err := m.cacher.Get(key)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 && ids[0] != "" {
				return ids[0], nil
			}
		}
	}
	listOfTeams, sndErr := m.teamsAPI.ListMyJoined(ctx)
	if sndErr != nil {
		return "", sndErr
	}
	id, err := resolveTeamIDByName(ref, listOfTeams)
	if err != nil {
		return "", err
	}
	if m.cacheEnabled && m.cacher != nil {
		key := cacher.NewTeamKeyBuilder(ref).ToString()
		_ = m.cacher.Set(key, id)
	}

	return id, nil
}

func resolveTeamIDByName(ref string, list msmodels.TeamCollectionResponseable) (string, error) {
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

func isLikelyGUID(s string) bool {
	var guidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return guidRegex.MatchString(s)
}
