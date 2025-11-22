package sender

import (
	"context"
	"errors"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

type GraphCall func(ctx context.Context) (*Response, error)

type RequestTechParams struct {
	MaxRetries     int
	NextRetryDelay int // in seconds
	Timeout        int // in seconds
}

func SendRequest(ctx context.Context, call GraphCall, techParams RequestTechParams) (*Response, *RequestError) { //TODO define response type
	var err error
	for attempt := 0; attempt < techParams.MaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(techParams.Timeout)*time.Second)
		defer cancel()
		response, err := call(ctx)
		if err == nil {
			//TODO log
			return response, nil
		}
		time.Sleep(time.Duration(techParams.NextRetryDelay) * time.Second)
	}

	//TODO log
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
