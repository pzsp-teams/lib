package mapper

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
)

type Mapper interface {
	MapTeamRefToTeamID(ctx context.Context, teamRef string) (string, error)
	MapChannelRefToChannelID(ctx context.Context, teamID, channelRef string) (string, error)
	MapUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error)
}

type mapper struct {
	teamsAPI    api.Teams
	channelsAPI api.Channels
}

func New(teamsAPI api.Teams, channelsAPI api.Channels) Mapper {
	return &mapper{
		teamsAPI:    teamsAPI,
		channelsAPI: channelsAPI,
	}
}

func (m *mapper) MapTeamRefToTeamID(ctx context.Context, teamRef string) (string, error) {
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
	if listOfTeams == nil || listOfTeams.GetValue() == nil || len(listOfTeams.GetValue()) == 0 {
		return "", fmt.Errorf("no teams available")
	}
	var matches []msmodels.Teamable
	for _, t := range listOfTeams.GetValue() {
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
			"multiple teams named %q found: %s. Please use one of the IDs instead",
			ref, strings.Join(options, "; "),
		)
	}
}

func (m *mapper) MapChannelRefToChannelID(ctx context.Context, teamID, channelRef string) (string, error) {
	ref := strings.TrimSpace(channelRef)
	if ref == "" {
		return "", fmt.Errorf("empty channel reference")
	}
	if isLikelyChannelID(ref) {
		return ref, nil
	}
	chans, err := m.channelsAPI.ListChannels(ctx, teamID)
	if err != nil {
		return "", err
	}
	if chans == nil || chans.GetValue() == nil || len(chans.GetValue()) == 0 {
		return "", fmt.Errorf("no channels available in team %q", teamID)
	}
	var matches []msmodels.Channelable
	for _, c := range chans.GetValue() {
		if c == nil {
			continue
		}
		if deref(c.GetDisplayName()) == ref {
			matches = append(matches, c)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("channel with name %q not found in team %q", ref, teamID)
	case 1:
		id := deref(matches[0].GetId())
		if id == "" {
			return "", fmt.Errorf("channel %q has nil id in team %q", ref, teamID)
		}
		return id, nil
	default:
		var options []string
		for _, c := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", deref(c.GetDisplayName()), deref(c.GetId())))
		}
		return "", fmt.Errorf(
			"multiple channels named %q found in team %q: %s. Please use one of the IDs instead",
			ref, teamID, strings.Join(options, "; "),
		)
	}
}

func (m *mapper) MapUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	resp, err := m.channelsAPI.ListMembers(ctx, teamID, channelID)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.GetValue() == nil || len(resp.GetValue()) == 0 {
		return "", fmt.Errorf("no members found in channel %q", channelID)
	}

	for _, member := range resp.GetValue() {
		if member == nil {
			continue
		}
		um, ok := member.(msmodels.AadUserConversationMemberable)
		if !ok {
			continue
		}
		if matchesUserRef(um, userRef) {
			return deref(member.GetId()), nil
		}
	}

	return "", fmt.Errorf("member %q not found in channel %q", userRef, channelID)
}

func matchesUserRef(um msmodels.AadUserConversationMemberable, userRef string) bool {
	if userRef == "" {
		return false
	}
	if deref(um.GetUserId()) == userRef {
		return true
	}
	if deref(um.GetDisplayName()) == userRef {
		return true
	}
	ad := um.GetAdditionalData()
	if ad == nil {
		return false
	}
	for _, key := range []string{"userPrincipalName", "mail", "email"} {
		raw, ok := ad[key]
		if !ok {
			continue
		}
		if v, ok := raw.(string); ok && strings.EqualFold(v, userRef) {
			return true
		}
	}
	return false
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

func isLikelyChannelID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@thread.")
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
