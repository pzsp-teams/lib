package models

// Team represents a Microsoft Teams team.
type Team struct {
	ID          string
	DisplayName string
	Description string
	Visibility  string
	IsArchived  bool
}
