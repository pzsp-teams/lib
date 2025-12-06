package cacher

type KeyType string
const (
	TeamName KeyType = "team"
	ChannelName KeyType = "channel"
)

type Key struct {
	Type KeyType
	Name string
} 

func (key *Key) ToString() string{
	t := string(key.Type)
	return "$" + t + "$:" + key.Name
}