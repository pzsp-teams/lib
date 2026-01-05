package models

// Team represents a Microsoft Teams team.
type Team struct {
	ID          string
	DisplayName string
	Description string
	IsArchived  bool
	Visibility  *string
}

// TeamUpdate represents the fields that can be updated for a Team.
type TeamUpdate struct {
	DisplayName *string
	Description *string
	Visibility  *string
}