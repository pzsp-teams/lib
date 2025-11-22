package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache"
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	authorityURL = "https://login.microsoftonline.com/"
)

var ErrUserNotFound = errors.New("user not found in MSAL cache")

type MSALTokenProvider struct {
	client *public.Client
	scopes []string
}

type MSALCredentials struct {
	ClientID string
	Tenet    string
	Scopes   []string
}

func NewMSALTokenProvider(credentials *MSALCredentials) (*MSALTokenProvider, error) {
	storage, err := accessor.New(credentials.ClientID)
	if err != nil {
		return nil, fmt.Errorf("creating persistant storage: %w", err)
	}

	cacheAccessor, err := cache.New(storage, credentials.ClientID)
	if err != nil {
		return nil, fmt.Errorf("creating cache: %w", err)
	}

	authority := authorityURL + credentials.Tenet
	client, err := public.New(
		credentials.ClientID,
		public.WithAuthority(authority),
		public.WithCache(cacheAccessor),
	)
	if err != nil {
		return nil, fmt.Errorf("creating MSALTokenProvider: %w", err)
	}

	return &MSALTokenProvider{client: &client, scopes: credentials.Scopes}, nil
}

func (p *MSALTokenProvider) GetToken(email string) (*AccessToken, error) {
	var userFound bool = false
	var result public.AuthResult

	accounts, err := p.client.Accounts(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("fetching cached accounts: %w", err)
	}

	if len(accounts) > 0 {
		if acc, err := resolveAccount(email, accounts); err == nil {
			if result, err = p.client.AcquireTokenSilent(
				context.TODO(),
				p.scopes,
				public.WithSilentAccount(*acc),
			); err == nil {
				userFound = true
			}
		}
	}

	if !userFound || len(accounts) == 0 {
		result, err = p.client.AcquireTokenInteractive(context.TODO(), p.scopes, public.WithLoginHint(email))
		if err != nil {
			return nil, fmt.Errorf("acquiring token: %w", err)
		}
	}

	return &AccessToken{
		AccessToken:   result.AccessToken,
		GrantedScopes: result.GrantedScopes,
		Expiry:        result.ExpiresOn,
	}, nil
}

func resolveAccount(email string, accounts []public.Account) (*public.Account, error) {
	for _, acc := range accounts {
		if acc.PreferredUsername == email {
			return &acc, nil
		}
	}
	return nil, ErrUserNotFound
}
