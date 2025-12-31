// Package models contains simplified Microsoft Teams domain types used by this library.
package models

/*
ChatType represents the type of chat in Microsoft Teams.
Type can be either one-on-one or group chat
*/
type ChatType string

const (
	// ChatTypeOneOnOne represents a one-on-one chat.
	ChatTypeOneOnOne ChatType = "one-on-one"
	// ChatTypeGroup represents a group chat.
	ChatTypeGroup ChatType = "group"
)

// Chat represents a chat in Microsoft Teams.
type Chat struct {
	ID       string
	Type     ChatType
	IsHidden bool
	Topic    *string
}
