package models

// Member represents a member of a Microsoft Teams channel, direct chat or team.
type Member struct {
	ID          string
	UserID      string
	DisplayName string
	Role        string
	Email       string
	ThreadID    *string
	TeamID      *string
}
