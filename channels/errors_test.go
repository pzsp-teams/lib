package channels

import (
	"errors"
	"strings"
	"testing"

	"github.com/pzsp-teams/lib/internal/sender"
)

func TestMapError_ResourceNotFound(t *testing.T) {
	e := &sender.RequestError{
		Code:    "ResourceNotFound",
		Message: "not found",
	}
	err := mapError(e)
	if !errors.Is(err, ErrChannelNotFound) {
		t.Fatalf("expected ErrChannelNotFound, got %v", err)
	}
}

func TestMapError_AccessDenied(t *testing.T) {
	e := &sender.RequestError{
		Code:    "AccessDenied",
		Message: "access denied",
	}
	err := mapError(e)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrChannelAccessDenied, got %v", err)
	}
}

func TestMapError_Unknown(t *testing.T) {
	e := &sender.RequestError{
		Code:    "SomeOtherError",
		Message: "some other error",
	}
	err := mapError(e)
	if !errors.Is(err, ErrUnknown) {
		t.Fatalf("expected ErrUnknown, got %v", err)
	}

	msg := err.Error()
	if !strings.Contains(msg, "some other error") || !strings.Contains(msg, "SomeOtherError") {
		t.Fatalf("error message does not contain original error details: %s", msg)
	}
}
