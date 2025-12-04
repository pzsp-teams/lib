package teams

import (
	"errors"
	"fmt"

	"github.com/pzsp-teams/lib/internal/sender"
)

var (
	errTeamNotFound = errors.New("team not found")
	errForbidden    = errors.New("forbidden access to team")
	errUnknown      = errors.New("unknown team error")
	errNotFound     = errors.New("not found")
)

func mapError(e *sender.RequestError) error {
	switch e.Code {
	case "ResourceNotFound":
		return errTeamNotFound
	case "AccessDenied":
		return errForbidden
	case "NotFound":
		return fmt.Errorf("%w: %s", errNotFound, e.Message)
	default:
		return fmt.Errorf("%w: %s (%s)", errUnknown, e.Message, e.Code)
	}
}
