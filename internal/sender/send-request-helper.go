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

func withTimeout(timeout time.Duration, call GraphCall) GraphCall {
	return func(ctx context.Context) (Response, error) {
		cctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return call(cctx)
	}
}

func retry(ctx context.Context, attempts int, delay time.Duration, call GraphCall) (Response, error) {
	var err error
	var res Response
	for i := 0; i < attempts; i++ {
		res, err = call(ctx)
		if err == nil {
			return res, nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return nil, err
}

// SendRequest will be used later by other packages
func SendRequest(ctx context.Context, call GraphCall, techParams RequestTechParams) (Response, *RequestError) {
	timeout := time.Duration(techParams.Timeout) * time.Second
	delay := time.Duration(techParams.NextRetryDelay) * time.Second
	call = withTimeout(timeout, call)
	res, err := retry(ctx, techParams.MaxRetries, delay, call)
	if err != nil {
		what.Happens("ERROR", "Request failed")
		return nil, convertGraphError(err)
	}
	what.Happens("INFO", "Request successful")
	what.Is(res)
	return res, nil
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
