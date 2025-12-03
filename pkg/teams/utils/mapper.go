package utils

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/pkg/teams/channels"
	"github.com/pzsp-teams/lib/pkg/teams/teams"
)

// MapperInterface will be used later
type MapperInterface interface {
	MapTeamNameToTeamID(teamName string) (string, error)
	MapChannelNameToChannelID(teamID, channelName string) (string, error)
}

// Mapper will be used later
type Mapper struct {
	teamSvc    teams.Service
	channelSvc channels.Service
}

// NewMapper will be used later
func NewMapper(teamSvc teams.Service, channelSvc channels.Service) *Mapper {
	return &Mapper{
		teamSvc:    teamSvc,
		channelSvc: channelSvc,
	}
}

// MapTeamNameToTeamID will be used later
func (m *Mapper) MapTeamNameToTeamID(teamName string) (string, error) {
	listOfTeams, err := m.teamSvc.ListMyJoined(context.TODO())
	if err != nil {
		return "", err
	}
	for _, t := range listOfTeams {
		if t.DisplayName == teamName {
			return t.ID, nil
		}
	}
	return "", fmt.Errorf("team with name '%s' not found", teamName)
}

// MapChannelNameToChannelID will be used later
func (m *Mapper) MapChannelNameToChannelID(teamID, channelName string) (string, error) {
	chans, err := m.channelSvc.ListChannels(context.TODO(), teamID)
	if err != nil {
		return "", err
	}
	for _, c := range chans {
		if c.Name == channelName {
			return c.ID, nil
		}
	}
	return "", fmt.Errorf("channel with name '%s' not found", channelName)
}
