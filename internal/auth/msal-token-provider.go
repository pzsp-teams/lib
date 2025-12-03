package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache"
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	authorityURL = "https://login.microsoftonline.com/"
	cacheDIR     = ".cache"
)

var errUserNotFound = errors.New("user not found in MSAL cache")

// Method will be used later by other packages
type Method string

const (
	interactive Method = "INTERACTIVE"
	deviceCode  Method = "DEVICE_CODE"
)

// MSALTokenProvider will be used later by other packages
type MSALTokenProvider struct {
	client     *public.Client
	authMethod Method
	scopes     []string
}

// MSALCredentials will be used later by other packages
type MSALCredentials struct {
	ClientID   string
	Tenant     string
	Scopes     []string
	AuthMethod Method
}

// NewMSALTokenProvider will be used later by other packages
func NewMSALTokenProvider(credentials *MSALCredentials) (*MSALTokenProvider, error) {
	storage, err := accessor.New(credentials.ClientID)
	if err != nil {
		return nil, fmt.Errorf("creating persistent storage: %w", err)
	}

	if err := os.MkdirAll(cacheDIR, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	cacheAccessor, err := cache.New(storage, cacheDIR+"/"+credentials.ClientID)
	if err != nil {
		return nil, fmt.Errorf("creating cache: %w", err)
	}

	authority := authorityURL + credentials.Tenant
	client, err := public.New(
		credentials.ClientID,
		public.WithAuthority(authority),
		public.WithCache(cacheAccessor),
	)
	if err != nil {
		return nil, fmt.Errorf("creating MSALTokenProvider: %w", err)
	}

	return &MSALTokenProvider{
		client:     &client,
		authMethod: credentials.AuthMethod,
		scopes:     credentials.Scopes,
	}, nil
}

// GetToken will be used later by other packages
func (p *MSALTokenProvider) GetToken(email string) (*AccessToken, error) {
	var result public.AuthResult
	var userFound bool

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
		switch p.authMethod {
		case interactive:
			result, err = p.client.AcquireTokenInteractive(
				context.TODO(),
				p.scopes,
				public.WithLoginHint(email),
			)
			if err != nil {
				return nil, fmt.Errorf("acquiring token interactively: %w", err)
			}
		case deviceCode:
			deviceCode, err := p.client.AcquireTokenByDeviceCode(
				context.TODO(),
				p.scopes,
			)
			if err != nil {
				return nil, fmt.Errorf("starting device code flow: %w", err)
			}
			fmt.Println(deviceCode.Result.Message)

			result, err = deviceCode.AuthenticationResult(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("completing device code auth: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported auth method: %s", p.authMethod)
		}
	}

	return &AccessToken{
		AccessToken:   result.AccessToken,
		GrantedScopes: result.GrantedScopes,
		Expiry:        result.ExpiresOn,
	}, nil
}

func resolveAccount(email string, accounts []public.Account) (*public.Account, error) {
	for i := range accounts {
		if accounts[i].PreferredUsername == email {
			return &accounts[i], nil
		}
	}
	return nil, errUserNotFound
}
