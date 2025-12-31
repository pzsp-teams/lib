package channels

import (
	"strings"

	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/models"
)

func tryAddTeamOrChannelMention(a *mentions.MentionAdder, raw, teamRef, teamID, channelRef, channelID string) bool {
	low := strings.ToLower(raw)

	if isTeamRef(low, raw, teamRef, teamID) {
		a.Add(models.MentionTeam, teamID, teamRef)
		return true
	}
	if isChannelRef(low, raw, channelRef, channelID) {
		a.Add(models.MentionChannel, channelID, channelRef)
		return true
	}
	return false
}

func isTeamRef(low, raw, teamRef, teamID string) bool {
	return low == "team" || raw == teamRef || raw == teamID
}

func isChannelRef(low, raw, channelRef, channelID string) bool {
	return low == "channel" || raw == channelRef || raw == channelID
}

func memberRole(isOwner bool) string {
	if isOwner {
		return "owner"
	}
	return "member"
}
