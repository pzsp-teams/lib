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

	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/sender"
)

type resolverContext[T any] struct {
	cacheKey    string
	ref         string
	isAlreadyID func() bool
	fetch       func(ctx context.Context) (T, *sender.RequestError)
	extract     func(data T) (string, error)
}

func (r *resolverContext[T]) resolveWithCache(ctx context.Context, c cacher.Cacher, cacheEnabled bool) (string, error) {
	if r.ref == "" {
		return "", fmt.Errorf("empty ref")
	}

	if r.isAlreadyID() {
		return r.ref, nil
	}

	if cacheEnabled && c != nil {
		value, found, err := c.Get(r.cacheKey)
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

	if cacheEnabled && c != nil {
		_ = c.Set(r.cacheKey, id)
	}

	return id, nil
}
