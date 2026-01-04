package auth

import (
	"fmt"

	"github.com/pzsp-teams/lib/config"
)

var (
	msalTokenProviderSingleton *MSALTokenProvider
)

func GetMSALTokenProvider(cfg *config.AuthConfig) (*MSALTokenProvider, error) {
	if msalTokenProviderSingleton == nil {
		provider, err := NewMSALTokenProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("creating MSAL token provider: %w", err)
		}
		msalTokenProviderSingleton = provider
	}
	return msalTokenProviderSingleton, nil
}