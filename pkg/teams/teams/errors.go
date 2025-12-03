package teams

import (
	"errors"
	"fmt"

	"github.com/pzsp-teams/lib/internal/sender"
)

var (
	ErrTeamNotFound = errors.New("team not found")
	ErrForbidden    = errors.New("forbidden access to team")
	ErrUnknown      = errors.New("unknown team error")
	ErrNotFound     = errors.New("not found")
)

func mapError(e *sender.RequestError) error {
	switch e.Code {
	case "ResourceNotFound":
		return ErrTeamNotFound
	case "AccessDenied":
		return ErrForbidden
	case "NotFound":
		return fmt.Errorf("%w: %s", ErrNotFound, e.Message)
	default:
		return fmt.Errorf("%w: %s (%s)", ErrUnknown, e.Message, e.Code)
	}
}
