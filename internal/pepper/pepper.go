// Package pepper provides helpers for obtaining and persisting a secret "pepper" value.
//
// The pepper is stored in the system keyring and is used as an additional secret when
// hashing or deriving cache keys. If the value is missing, the user is prompted on stdin
// and the result is saved to the keyring.
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

var ErrPepperNotSet = errors.New("pepper not set in keyring")

var keyringGet = keyring.Get
var keyringSet = keyring.Set

func GetPepper() (string, error) {
	value, err := keyringGet(serviceName, userName)
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

func SetPepper(pepper string) error {
	pepper = strings.TrimSpace(pepper)
	if pepper == "" {
		return fmt.Errorf("pepper cannot be empty")
	}
	if err := keyringSet(serviceName, userName, pepper); err != nil {
		return fmt.Errorf("storing pepper in keyring: %w", err)
	}
	return nil
}

func PepperExists() (bool, error) {
	value, err := keyringGet(serviceName, userName)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("checking pepper in keyring: %w", err)
	}
	return strings.TrimSpace(value) != "", nil
}