package channels

import (
	"errors"
	"fmt"

	"github.com/pzsp-teams/lib/internal/sender"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrForbidden      = errors.New("forbidden access to channel")
	ErrUnknown        = errors.New("unknown channel error")
)

func mapError(e *sender.RequestError) error {
	switch e.Code {
	case "ResourceNotFound":
		return ErrChannelNotFound
	case "AccessDenied":
		return ErrForbidden
	default:
		return fmt.Errorf("%w: %s (%s)", ErrUnknown, e.Message, e.Code)
	}
}