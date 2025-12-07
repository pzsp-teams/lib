package cacher

import "fmt"

type KeyType string
const (
	Team KeyType = "team"
	Channel KeyType = "channel"
)

type KeyBuilder interface{
	ToString() string
}

type TeamKeyBuilder struct {
	Type KeyType
	Name string
} 

func NewTeamKeyBuilder(name string) (KeyBuilder) {
	return &TeamKeyBuilder{
		Type: Team,
		Name: name,
	}
}

type ChannelKeyBuilder struct {
	Type KeyType
	TeamID string
	Name string
}

func NewChannelKeyBuilder(teamID, name string) (KeyBuilder) {
	return &ChannelKeyBuilder{
		Type: Channel,
		TeamID: teamID,
		Name: name,
	}
}

func (keyBuilder *TeamKeyBuilder) ToString() string{
	return fmt.Sprintf("$%v$:%v", keyBuilder.Type, keyBuilder.Name)
}

func (keyBuilder *ChannelKeyBuilder) ToString() string{
	return fmt.Sprintf("$%v$:%v:%v", keyBuilder.Type, keyBuilder.TeamID,keyBuilder.Name)
}