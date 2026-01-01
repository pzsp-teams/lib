package models

// Channel represents a Microsoft Teams channel.
type Channel struct {
	TeamID    string
	ID        string
	Name      string
	IsGeneral bool
}
