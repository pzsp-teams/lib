package auth

import "time"

// AccessToken will be used later by other packages
type AccessToken struct {
	AccessToken   string
	GrantedScopes []string
	Expiry        time.Time
}

// TokenProvider will be used later by other packages
type TokenProvider interface {
	GetToken(email string) (*AccessToken, error)
}
