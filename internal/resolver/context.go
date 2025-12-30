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

// ResolverContext encapsulates the context and logic needed to resolve a reference into a resource ID.
// It is a generic struct that can work with any type T representing the data fetched from the API.
type ResolverContext[T any] struct {
	cacheKey    string
	keyType     cacher.KeyType
	ref         string
	isAlreadyID func() bool
	fetch       func(ctx context.Context) (T, *sender.RequestError)
	extract     func(data T) (string, error)
}

// resolveWithCache resolves the reference into a Microsoft Graph resource ID, utilizing caching if enabled.
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
