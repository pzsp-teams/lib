package teams

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/pzsp-teams/lib/internal/auth"
)

// AuthConfig will be used later
type AuthConfig struct {
	ClientID   string
	Tenant     string
	Scopes     []string
	AuthMethod string
	Email      string
}

type msalCredential struct {
	provider auth.TokenProvider
	email    string
}

// GetToken will be used later
func (c *msalCredential) GetToken(ctx context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	token, err := c.provider.GetToken(c.email)
	if err != nil {
		return azcore.AccessToken{}, err
	}
	return azcore.AccessToken{
		Token:     token.AccessToken,
		ExpiresOn: token.Expiry,
	}, nil
}
