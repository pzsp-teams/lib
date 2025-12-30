package models

type MentionKind string

const (
	MentionUser MentionKind = "user"
	MentionChannel MentionKind = "channel"
	MentionTeam MentionKind = "team"
	MentionEveryone MentionKind = "everyone"
)

type Mention struct {
	Kind 	  	MentionKind
	AtID        int32
	Text 		string
	TargetID    string
}