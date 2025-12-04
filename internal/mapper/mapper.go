package mapper

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/internal/api"
)

// Mapper will be used later
type Mapper interface {
	MapTeamNameToTeamID(ctx context.Context, teamName string) (string, error)
	MapChannelNameToChannelID(ctx context.Context, teamID, channelName string) (string, error)
}

type mapper struct {
	teamsAPI    api.Teams
	channelsAPI api.Channels
}

// New will be used later
func New(teamsAPI api.Teams, channelsAPI api.Channels) Mapper {
	return &mapper{
		teamsAPI:    teamsAPI,
		channelsAPI: channelsAPI,
	}
}

// MapTeamNameToTeamID will be used later
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

// MapChannelNameToChannelID will be used later
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
