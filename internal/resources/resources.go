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
	Mention       Resource = "MENTION"
)

type Key string

const (
	TeamRef         Key = "team_ref"
	ChannelRef      Key = "channel_ref"
	ChatRef         Key = "chat_ref"
	GroupChatRef    Key = "group_chat_ref"
	OneOnOneChatRef Key = "one_on_one_chat_ref"
	UserRef         Key = "user_ref"
	MessageRef      Key = "message_ref"
	MentionRef      Key = "mention_ref"
)
