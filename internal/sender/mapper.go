package sender

import (
	"net/http"

	"github.com/pzsp-teams/lib/internal/resources"
)

type Option func(*ErrData)

func WithResource(resType resources.Resource, resRef string) Option {
	return func(data *ErrData) {
		if data.ResourceRefs == nil {
			data.ResourceRefs = make(map[resources.Resource]string)
		}
		data.ResourceRefs[resType] = resRef
	}
}

func WithResources(resType resources.Resource, resRefs []string) Option {
	return func(data *ErrData) {
		if data.ResourceRefs == nil {
			data.ResourceRefs = make(map[resources.Resource]string)
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
		return ErrAccessForbidden{e.Code, e.Message, *data}

	case http.StatusNotFound:
		return ErrResourceNotFound{e.Code, e.Message, *data}

	default:
		return *e
	}
}
