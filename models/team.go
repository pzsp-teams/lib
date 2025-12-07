package models

// Team represents a Microsoft Teams team.
type Team struct {
	ID          string
	DisplayName string
	Description string
	IsArchived  bool
	Visibility  *string
}
