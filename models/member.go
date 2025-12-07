package models

// Member represents a member of a Microsoft Teams channel or direct chat.
type Member struct {
	ID          string
	UserID      string
	DisplayName string
	Role        string
}
