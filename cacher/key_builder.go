package cacher

import (
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/pepper"
	"github.com/pzsp-teams/lib/internal/util"
)

type KeyType string

const (
	Team    KeyType = "team"
	Channel KeyType = "channel"
	Member  KeyType = "member"
)

func formatKey(t KeyType, parts ...string) string {
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return fmt.Sprintf("$%s$:%s", t, strings.Join(parts, ":"))
}

func NewTeamKey(name string) string {
	return formatKey(Team, name)
}

func NewChannelKey(teamID, name string) string {
	return formatKey(Channel, teamID, name)
}

func NewMemberKey(ref, teamID, channelID string, pep *string) string {
	if pep == nil {
		p, err := pepper.GetOrAskPepper()
		if err != nil {
			p = "default-pepper"
		}
		pep = &p
	}
	hashedRef := util.HashWithPepper(*pep, ref)
	return formatKey(Member, teamID, channelID, hashedRef)
}
