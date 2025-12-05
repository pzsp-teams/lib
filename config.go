package lib

import (
	"github.com/pzsp-teams/lib/internal/auth"
	"github.com/pzsp-teams/lib/internal/sender"
)

type SenderConfig struct {
	MaxRetries     int
	NextRetryDelay int
	Timeout        int
}

func (cfg *SenderConfig) toTechParams() sender.RequestTechParams {
	return sender.RequestTechParams{
		MaxRetries:     cfg.MaxRetries,
		NextRetryDelay: cfg.NextRetryDelay,
		Timeout:        cfg.Timeout,
	}
}

type AuthConfig struct {
	ClientID   string
	Tenant     string
	Email      string
	Scopes     []string
	AuthMethod string
}

func (cfg *AuthConfig) toMSALCredentials() *auth.MSALCredentials {
	return &auth.MSALCredentials{
		ClientID:   cfg.ClientID,
		Tenant:     cfg.Tenant,
		Email:      cfg.Email,
		Scopes:     cfg.Scopes,
		AuthMethod: auth.Method(cfg.AuthMethod),
	}
}
