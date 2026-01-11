// Package resolver provides helpers for resolving user-facing references into Microsoft Graph resource IDs.
// It supports caching of resolved IDs to improve performance and reduce API calls.
//
// Resources that can be resolved: teams, channels, channel-members, one-on-one chats, group chats and chat-members.
//
// Sometimes resolver cannot unambiguously resolve a reference (e.g., multiple chats with the same topic).
// In such cases, it returns an error indicating the ambiguity.
package resolver

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/sender"
)

type resolverContext[T any] struct {
	cacheKey    string
	ref         string
	isAlreadyID func() bool
	fetch       func(ctx context.Context) (T, *sender.RequestError)
	extract     func(data T) (string, error)
}

func (r *resolverContext[T]) resolveWithCache(
	ctx context.Context,
	cacheHandler *cacher.CacheHandler,
) (string, error) {
	if r.ref == "" {
		return "", fmt.Errorf("empty ref")
	}

	if r.isAlreadyID() {
		return r.ref, nil
	}

	if cacheHandler != nil {
		value, found, err := cacheHandler.Cacher.Get(r.cacheKey)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 {
				return ids[0], nil
			} else if ok && len(ids) > 1 {
				cacheHandler.Runner.Run(func() {
					_ = cacheHandler.Cacher.Invalidate(r.cacheKey)
				})
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

	if cacheHandler != nil {
		cacheHandler.Runner.Run(func() {
			_ = cacheHandler.Cacher.Set(r.cacheKey, id)
		})
	}

	return id, nil
}
