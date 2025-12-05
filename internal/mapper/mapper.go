package mapper

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
)

type Mapper interface {
	MapTeamNameToTeamID(ctx context.Context, teamName string) (string, error)
	MapChannelNameToChannelID(ctx context.Context, teamID, channelName string) (string, error)
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

func (m *mapper) MapTeamNameToTeamID(ctx context.Context, teamName string) (string, error) {
	listOfTeams, err := m.teamsAPI.ListMyJoined(ctx)
	if err != nil {
		return "", err
	}
	if listOfTeams == nil || listOfTeams.GetValue() == nil {
		return "", fmt.Errorf("no teams available")
	}
	for _, t := range listOfTeams.GetValue() {
		if t.GetDisplayName() != nil && *t.GetDisplayName() == teamName {
			if t.GetId() == nil {
				return "", fmt.Errorf("team %q has nil id", teamName)
			}
			return *t.GetId(), nil
		}
	}
	return "", fmt.Errorf("team with name %q not found", teamName)
}

func (m *mapper) MapChannelNameToChannelID(ctx context.Context, teamID, channelName string) (string, error) {
	chans, err := m.channelsAPI.ListChannels(ctx, teamID)
	if err != nil {
		return "", err
	}
	if chans == nil || chans.GetValue() == nil {
		return "", fmt.Errorf("no channels available in team %q", teamID)
	}
	for _, c := range chans.GetValue() {
		if c.GetDisplayName() != nil && *c.GetDisplayName() == channelName {
			if c.GetId() == nil {
				return "", fmt.Errorf("channel %q has nil id in team %q", channelName, teamID)
			}
			return *c.GetId(), nil
		}
	}
	return "", fmt.Errorf("channel with name %q not found in team %q", channelName, teamID)
}

func (m *mapper) MapUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	resp, err := m.channelsAPI.ListMembers(ctx, teamID, channelID)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.GetValue() == nil {
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

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
