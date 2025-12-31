package config

// SenderConfig defines configuration for the request sender
// which connects with the Microsoft Graph API.
// MaxRetryDelay and Timeout are in seconds.
type SenderConfig struct {
	MaxRetries     int
	NextRetryDelay int
	Timeout        int
}
