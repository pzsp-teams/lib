package resolver

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/util"
)

// TeamResolver will be used later
type TeamResolver interface {
	ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error)
}

type teamResolver struct {
	teamsAPI api.TeamAPI
}

// New will be used later
func NewTeamResolver(teamsAPI api.TeamAPI, channelsAPI api.ChannelAPI) TeamResolver {
	return &teamResolver{
		teamsAPI: teamsAPI,
	}
}

// MapTeamNameToTeamID will be used later
func (m *teamResolver) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	ref := strings.TrimSpace(teamRef)
	if ref == "" {
		return "", fmt.Errorf("empty team reference")
	}
	if isLikelyGUID(ref) {
		return ref, nil
	}
	listOfTeams, err := m.teamsAPI.ListMyJoined(ctx)
	if err != nil {
		return "", err
	}
	return resolveTeamIDByName(ref, listOfTeams)
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
