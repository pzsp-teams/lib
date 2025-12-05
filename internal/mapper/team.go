package mapper

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
)

// TeamMapper will be used later
type TeamMapper interface {
	MapTeamRefToTeamID(ctx context.Context, teamRef string) (string, error)
}

type teamMapper struct {
	teamsAPI api.TeamAPI
}

// New will be used later
func NewTeamMapper(teamsAPI api.TeamAPI, channelsAPI api.ChannelAPI) TeamMapper {
	return &teamMapper{
		teamsAPI: teamsAPI,
	}
}

// MapTeamNameToTeamID will be used later
func (m *teamMapper) MapTeamRefToTeamID(ctx context.Context, teamRef string) (string, error) {
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
		if deref(t.GetDisplayName()) == ref {
			matches = append(matches, t)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("team with name %q not found", ref)
	case 1:
		id := deref(matches[0].GetId())
		if id == "" {
			return "", fmt.Errorf("team %q has nil id", ref)
		}
		return id, nil
	default:
		var options []string
		for _, t := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", deref(t.GetDisplayName()), deref(t.GetId())))
		}
		return "", fmt.Errorf(
			"multiple teams named %q found: \n%s.\nPlease use one of the IDs instead",
			ref, strings.Join(options, ";\n"),
		)
	}
}

func isLikelyGUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !isHexDigit(r) {
				return false
			}
		}
	}
	return true
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'f') ||
		(r >= 'A' && r <= 'F')
}
