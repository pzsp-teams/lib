package sender

import (
	"errors"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

func TestConvertGraphError_ODataError(t *testing.T) {
	odataErr := odataerrors.NewODataError()
	mainErr := odataerrors.NewMainError()

	code := "AccessDenied"
	msg := "Access is denied"

	mainErr.SetCode(&code)
	mainErr.SetMessage(&msg)

	odataErr.SetErrorEscaped(mainErr)

	got := convertGraphError(odataErr)

	if got.Code != code {
		t.Errorf("expected code %q, got %q", code, got.Code)
	}
	if got.Message != msg {
		t.Errorf("expected message %q, got %q", msg, got.Message)
	}
}

func TestConvertGraphError_GenericError(t *testing.T) {
	base := errors.New("boom")
	got := convertGraphError(base)

	if got.Code != "ParsingError" {
		t.Errorf("expected code ParsingError, got %q", got.Code)
	}
	if got.Message != "boom" {
		t.Errorf("expected message 'boom', got %q", got.Message)
	}
}

func TestConvertGraphError_NilError(t *testing.T) {
	var base error = nil
	got := convertGraphError(base)

	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
