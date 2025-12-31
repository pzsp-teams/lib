// Package auth provides an Azure AD (MSAL) token provider implementation
// backed by a persistent MSAL cache. It supports silent authentication with
// fallback to interactive or device-code flows.
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache"
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/pzsp-teams/lib/config"
)

const (
	authorityURL = "https://login.microsoftonline.com/"
)

var errUserNotFound = errors.New("user not found in MSAL cache")

// MSALTokenProvider implements azcore.TokenCredential using
// the Microsoft Authentication Library (MSAL).
// Tokens are acquired silently from cache when possible and
// fall back to the configured interactive or device-code flow.
type MSALTokenProvider struct {
	client     *public.Client
	email      string
	scopes     []string
	authMethod config.Method
}

// NewMSALTokenProvider creates a token provider backed by a persistent
// MSAL cache stored in the user's cache directory.
func NewMSALTokenProvider(cfg *config.AuthConfig) (*MSALTokenProvider, error) {
	storage, err := accessor.New(cfg.ClientID)
	if err != nil {
		return nil, fmt.Errorf("creating persistent storage: %w", err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("retrieving cache dir: %w", err)
	}

	cacheAccessor, err := cache.New(storage, filepath.Join(cacheDir, "pzsp-teams", cfg.ClientID))
	if err != nil {
		return nil, fmt.Errorf("creating cache: %w", err)
	}

	authority := authorityURL + cfg.Tenant
	client, err := public.New(
		cfg.ClientID,
		public.WithAuthority(authority),
		public.WithCache(cacheAccessor),
	)
	if err != nil {
		return nil, fmt.Errorf("creating msal client: %w", err)
	}

	return &MSALTokenProvider{
		client:     &client,
		email:      cfg.Email,
		scopes:     cfg.Scopes,
		authMethod: cfg.AuthMethod,
	}, nil
}

// GetToken implements azcore.TokenCredential.
//
// It attempts to acquire a token silently using cached accounts.
// If no matching account is found or silent acquisition fails,
// it falls back to the configured authentication flow.
func (p *MSALTokenProvider) GetToken(ctx context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	var result public.AuthResult
	var userFound bool
	nilToken := azcore.AccessToken{}

	accounts, err := p.client.Accounts(ctx)
	if err != nil {
		return nilToken, fmt.Errorf("fetching cached accounts: %w", err)
	}

	if len(accounts) > 0 {
		if acc, err := resolveAccount(p.email, accounts); err == nil {
			if result, err = p.client.AcquireTokenSilent(
				ctx,
				p.scopes,
				public.WithSilentAccount(*acc),
			); err == nil {
				userFound = true
			}
		}
	}

	if !userFound || len(accounts) == 0 {
		switch p.authMethod {
		case config.Interactive:
			result, err = p.client.AcquireTokenInteractive(
				ctx,
				p.scopes,
				public.WithLoginHint(p.email),
			)
			if err != nil {
				return nilToken, fmt.Errorf("acquiring token interactively: %w", err)
			}
		case config.DeviceCode:
			deviceCode, err := p.client.AcquireTokenByDeviceCode(
				ctx,
				p.scopes,
			)
			if err != nil {
				return nilToken, fmt.Errorf("starting device code flow: %w", err)
			}
			fmt.Println(deviceCode.Result.Message)

			result, err = deviceCode.AuthenticationResult(ctx)
			if err != nil {
				return nilToken, fmt.Errorf("completing device code auth: %w", err)
			}
		default:
			return nilToken, fmt.Errorf("unsupported auth method: %s", p.authMethod)
		}
	}

	return azcore.AccessToken{
		Token:     result.AccessToken,
		ExpiresOn: result.ExpiresOn,
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
