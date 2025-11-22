package auth

import "time"

type AccessToken struct {
	AccessToken   string
	GrantedScopes []string
	Expiry        time.Time
}

type TokenProvider interface {
	GetToken(email string) (*AccessToken, error)
}
