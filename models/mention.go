package models

// MentionKind represents the kind of mention in a Microsoft Teams message.
type MentionKind string

const (
	// MentionUser represents a user mention.
	MentionUser MentionKind = "user"
	// MentionChannel represents a channel mention.
	MentionChannel MentionKind = "channel"
	// MentionTeam represents a team mention.
	MentionTeam MentionKind = "team"
	// MentionEveryone represents an everyone mention - applicable to group chats only
	MentionEveryone MentionKind = "everyone"
)

// Mention represents a mention in a Microsoft Teams message.
type Mention struct {
	Kind     MentionKind
	AtID     int32
	Text     string
	TargetID string
}
