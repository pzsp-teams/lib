package sender

import (
	"net/http"
)

type Option func(*ErrData)

func WithResource(resType Resource, resRef string) Option {
	return func(data *ErrData) {
		if data.ResourceRefs == nil {
			data.ResourceRefs = make(map[Resource]string)
		}
		data.ResourceRefs[resType] = resRef
	}
}

func WithResources(resType Resource, resRefs []string) Option {
	return func(data *ErrData) {
		if data.ResourceRefs == nil {
			data.ResourceRefs = make(map[Resource]string)
		}
		for _, ref := range resRefs {
			data.ResourceRefs[resType] = ref
		}
	}
}

func MapError(e *RequestError, opts ...Option) error {
	data := &ErrData{}
	for _, opt := range opts {
		opt(data)
	}
	switch e.Code {
	case http.StatusForbidden:
		return ErrAccessForbidden{e.Code, *data}

	case http.StatusNotFound:
		return ErrResourceNotFound{e.Code, *data}

	default:
		return *e
	}
}
