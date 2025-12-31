// Package sender provides helpers for executing Microsoft Graph requests with retries and timeouts.
//
// It defines GraphCall and SendRequest, converts Graph/OData errors into library-specific error types,
// and can enrich errors with resource context (e.g. TEAM/CHANNEL/CHAT refs) for easier debugging.
package sender

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/pzsp-teams/lib/config"
)

type GraphCall func(ctx context.Context) (Response, error)

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
		if !shouldRetry(err) || i >= attempts-1 {
			break
		}
		time.Sleep(delay)
	}
	return nil, err
}

func SendRequest(ctx context.Context, call GraphCall, cfg *config.SenderConfig) (Response, *RequestError) {
	timeout := time.Duration(cfg.Timeout) * time.Second
	delay := time.Duration(cfg.NextRetryDelay) * time.Second
	call = withTimeout(timeout, call)
	res, err := retry(ctx, cfg.MaxRetries, delay, call)
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
