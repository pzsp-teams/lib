// Package pepper provides helpers for persisting and retrieving a secret "pepper" value.
//
// The pepper is stored in the system keyring and is used as an additional secret when
// hashing or deriving cache keys.
//
// This package is intentionally non-interactive: it never prompts the user.
// Use SetPepper to store the value, PepperExists to check if it is present,
// and GetPepper to retrieve it.
package pepper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "pzsp-teams-cache"
	userName    = "pepper"
)

// ErrPepperNotSet is returned by GetPepper when the pepper is missing or empty in the keyring.
var ErrPepperNotSet = errors.New("pepper not set in keyring")

// GetPepper retrieves the pepper from the system keyring.
//
// It returns ErrPepperNotSet if the value is missing or empty.
func GetPepper() (string, error) {
	value, err := keyring.Get(serviceName, userName)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrPepperNotSet
		}
		return "", fmt.Errorf("retrieving pepper from keyring: %w", err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrPepperNotSet
	}
	return value, nil
}

// SetPepper validates and stores the pepper in the system keyring.
func SetPepper(pepper string) error {
	pepper = strings.TrimSpace(pepper)
	if pepper == "" {
		return fmt.Errorf("pepper cannot be empty")
	}
	if err := keyring.Set(serviceName, userName, pepper); err != nil {
		return fmt.Errorf("storing pepper in keyring: %w", err)
	}
	return nil
}

// PepperExists reports whether a non-empty pepper is stored in the system keyring.
func PepperExists() (bool, error) {
	value, err := keyring.Get(serviceName, userName)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("checking pepper in keyring: %w", err)
	}
	return strings.TrimSpace(value) != "", nil
}
