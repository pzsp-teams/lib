package auth

import "time"

type AccessToken interface {
	Token() string
	ExpiresAt() time.Time
}

type TokenProvider interface {
	GetToken(email string) (AccessToken, error)
}
