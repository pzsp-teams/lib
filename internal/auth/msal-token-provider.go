package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	authorityURL = "https://login.microsoftonline.com/"
)

var ErrUserNotFound = errors.New("user not found in MSAL cache")

type MSALAccessToken struct {
	AccessToken string
	Expiry      time.Time
}

func (t *MSALAccessToken) Token() string {
	return t.AccessToken
}

func (t *MSALAccessToken) ExpiresAt() time.Time {
	return t.Expiry
}

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
	authority := authorityURL + credentials.Tenet
	client, err := public.New(credentials.ClientID, public.WithAuthority(authority))
	if err != nil {
		return nil, fmt.Errorf("creating MSALTokenProvider: %w", err)
	}

	return &MSALTokenProvider{client: &client, scopes: credentials.Scopes}, nil
}

func (p *MSALTokenProvider) GetToken(email string) (AccessToken, error) {
	var result public.AuthResult

	accounts, err := p.client.Accounts(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("fetching cached accounts: %w", err)
	}

	if len(accounts) > 0 {
		if acc, accErr := resolveAccount(email, accounts); accErr == nil {
			result, err = p.client.AcquireTokenSilent(context.TODO(), p.scopes, public.WithSilentAccount(*acc))
		}
	}

	if err != nil || len(accounts) == 0 {
		result, err = p.client.AcquireTokenInteractive(context.TODO(), p.scopes, public.WithLoginHint(email))
		if err != nil {
			return nil, fmt.Errorf("acquiring token interactively: %w", err)
		}
	}

	return &MSALAccessToken{AccessToken: result.AccessToken, Expiry: result.ExpiresOn}, nil
}

func resolveAccount(email string, accounts []public.Account) (*public.Account, error) {
	for _, acc := range accounts {
		if acc.PreferredUsername == email {
			return &acc, nil
		}
	}
	return nil, ErrUserNotFound
}
