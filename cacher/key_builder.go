package cacher

import "fmt"

type KeyType string
const (
	TeamName KeyType = "team"
	ChannelName KeyType = "channel"
)

type KeyBuilder interface{
	ToString() string
}

type TeamKeyBuilder struct {
	Type KeyType
	Name string
} 

type ChannelKeyBuilder struct {
	Type KeyType
	TeamID string
	Name string
}

func (keyBuilder *TeamKeyBuilder) ToString() string{
	return fmt.Sprintf("$%v$:%v", keyBuilder.Type, keyBuilder.Name)
}

func (keyBuilder *ChannelKeyBuilder) ToString() string{
	return fmt.Sprintf("$%v$:%v:%v", keyBuilder.Type, keyBuilder.TeamID,keyBuilder.Name)
}