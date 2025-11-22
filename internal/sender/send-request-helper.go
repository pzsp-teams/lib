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
	var err error
	for attempt := 0; attempt < techParams.MaxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(techParams.Timeout)*time.Second)
		response, err := call(attemptCtx)
		cancel()
		if err == nil {
			what.Happens("INFO", "Request successful")
			what.Is(response) // temp logs
			return response, nil
		}
		time.Sleep(time.Duration(techParams.NextRetryDelay) * time.Second)
	}

	what.Happens("ERROR", "SendRequest error") // temporary log so the package doesn't get removed
	return nil, convertGraphError(err)
}

func convertGraphError(err error) *RequestError {
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
