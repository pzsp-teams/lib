package sender

import (
	"context"
	"errors"
	"time"

	"appliedgo.net/what"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

// GraphCall will be used later by other packages
type GraphCall func(ctx context.Context) (Response, error)

// RequestTechParams will be used later by other packages
type RequestTechParams struct {
	MaxRetries     int
	NextRetryDelay int // in seconds
	Timeout        int // in seconds
}

// SendRequest will be used later by other packages
func SendRequest(ctx context.Context, call GraphCall, techParams RequestTechParams) (Response, *RequestError) {
	for attempt := 0; attempt < techParams.MaxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(techParams.Timeout)*time.Second)
		response, err := call(attemptCtx)
		cancel()
		if err == nil {
			what.Happens("INFO", "Request successful")
			what.Is(response) // temp logs
			return response, nil
		}
		if attempt == techParams.MaxRetries-1 {
			what.Happens("ERROR", "SendRequest error")
			return nil, convertGraphError(err)
		}
		time.Sleep(time.Duration(techParams.NextRetryDelay) * time.Second)
	}
	return nil, &RequestError{Code: "UnknownError", Message: "An unknown error occurred"}
}

func convertGraphError(err error) *RequestError {
	if err == nil {
		return nil
	}
	var odataErr *odataerrors.ODataError
	if errors.As(err, &odataErr) {
		errElapsed := odataErr.GetErrorEscaped()
		code := errElapsed.GetCode()
		message := errElapsed.GetMessage()
		return &RequestError{
			Code:    *code,
			Message: *message,
		}
	} else {
		return &RequestError{
			Code:    "ParsingError",
			Message: err.Error(),
		}
	}
}
