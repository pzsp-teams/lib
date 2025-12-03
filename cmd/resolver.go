package main

import (
	"strings"

	"github.com/pzsp-teams/lib/pkg/teams/utils"
)

func resolveTeamID(mapper *utils.Mapper, raw string) (string, error) {
	if looksLikeGUID(raw) {
		return raw, nil
	}
	return mapper.MapTeamNameToTeamID(raw)
}

func looksLikeGUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		case c == '-' && (i == 8 || i == 13 || i == 18 || i == 23):
		default:
			return false
		}
	}
	return true
}

func resolveChannelID(mapper *utils.Mapper, teamID, raw string) (string, error) {
	if strings.Contains(raw, "@thread.") || strings.HasPrefix(raw, "19:") {
		return raw, nil
	}
	return mapper.MapChannelNameToChannelID(teamID, raw)
}
