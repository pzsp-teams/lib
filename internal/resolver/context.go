package resolver

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/sender"
)

type ResolverContext[T any] struct {
	cacheKey    string
	keyType     cacher.KeyType
	ref         string
	isAlreadyID func() bool
	fetch       func(ctx context.Context) (T, *sender.RequestError)
	extract     func(data T) (string, error)
}

func (r *ResolverContext[T]) resolveWithCache(ctx context.Context, cacher cacher.Cacher, cacheEnabled bool) (string, error) {
	if r.ref == "" {
		return "", fmt.Errorf("empty ref")
	}

	if r.isAlreadyID() {
		return r.ref, nil
	}

	if cacheEnabled && cacher != nil {
		value, found, err := cacher.Get(r.cacheKey)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 {
				return ids[0], nil
			}
		}
	}

	data, apiErr := r.fetch(ctx)
	if apiErr != nil {
		return "", apiErr
	}

	id, err := r.extract(data)
	if err != nil {
		return "", err
	}

	if cacheEnabled && cacher != nil {
		_ = cacher.Set(r.cacheKey, id)
	}

	return id, nil
}
