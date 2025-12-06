package sender

import (
	"context"
	"errors"
	"net/http"
	"time"

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
	for i := range attempts {
		res, err = call(ctx)
		if err == nil {
			return res, nil
		}
		if !shouldRetry(err) || i < attempts-1 {
			break
		}
		time.Sleep(delay)
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
		return nil, convertGraphError(err)
	}
	return res, nil
}

func convertGraphError(err error) *RequestError {
	if err == nil {
		return nil
	}
	var odataErr *odataerrors.ODataError
	if errors.As(err, &odataErr) {
		errElapsed := odataErr.GetErrorEscaped()
		code := odataErr.GetStatusCode()
		message := errElapsed.GetMessage()
		return &RequestError{
			Code:    code,
			Message: *message,
		}
	} else {
		return &RequestError{
			Code:    http.StatusUnprocessableEntity,
			Message: err.Error(),
		}
	}
}

func shouldRetry(err error) bool {
	var odataErr *odataerrors.ODataError
	if errors.As(err, &odataErr) {
		if odataErr.GetStatusCode() > 500 {
			return true
		}
	}
	return false
}
