package api

import (
	"net/http"

	"github.com/pzsp-teams/lib/internal/sender"
)

const (
	graphUserBindFmt  = "https://graph.microsoft.com/v1.0/users('%s')"
	graphUserBindKey  = "user@odata.bind"
	templateBindKey   = "template@odata.bind"
	templateBindValue = "https://graph.microsoft.com/v1.0/teamsTemplates('standard')"
)

func newTypeError(expected string) *sender.RequestError {
	return &sender.RequestError{
		Code:    http.StatusUnprocessableEntity,
		Message: "Expected " + expected,
	}
}
