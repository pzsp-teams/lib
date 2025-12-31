// Package config holds configuration structs used across the application.
// Defined configs:
//   - AuthConfig: holds authentication configuration.
//   - SenderConfig: holds sender configuration.
//   - CacheConfig: holds caching configuration.
package config

// Method defines the authentication flow used when	acquiring tokens.
type Method string

const (
	// Interactive opens a browser window for user authentication.
	Interactive Method = "INTERACTIVE"

	// DeviceCode prints a device code to the console and prompts the user to visit a URL to authenticate.
	DeviceCode Method = "DEVICE_CODE"
)

// AuthConfig holds configuration for authentication.
// All fields are required - they are needed to acquire tokens via MSAL.
type AuthConfig struct {
	ClientID   string
	Tenant     string
	Email      string
	Scopes     []string
	AuthMethod Method
}
