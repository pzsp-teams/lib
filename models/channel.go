package models

// Channel represents a Microsoft Teams channel.
type Channel struct {
	ID        string
	TeamID    string
	Name      string
	IsGeneral bool
}
