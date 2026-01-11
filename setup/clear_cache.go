// Package setup provides setup utilities for the application.
// It includes functionalities for managing peppers and clearing caches.
package setup

import (
	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/cacher"
)

// ClearCache clears the cache based on the provided cache configuration.
func ClearCache(cacheCfg *config.CacheConfig) error {
	cacheHandler := cacher.GetCacheHandler(cacheCfg)
	if cacheHandler == nil {
		return nil
	}
	return cacheHandler.Cacher.Clear()
}
