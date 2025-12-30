package resources

type Resource string

const (
	Team          Resource = "TEAM"
	Channel       Resource = "CHANNEL"
	Chat          Resource = "CHAT"
	GroupChat     Resource = "GROUP_CHAT"
	OneOnOneChat  Resource = "ONE_ON_ONE_CHAT"
	User          Resource = "USER"
	Message       Resource = "MESSAGE"
	PinnedMessage Resource = "PINNED_MESSAGE"
)
