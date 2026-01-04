package cacher

import "github.com/pzsp-teams/lib/config"

var Singleton *CacheHandler

func GetCacheHandler(cfg *config.CacheConfig) *CacheHandler {
	if Singleton == nil {
		Singleton = NewCacheHandler(cfg)
	}
	return Singleton
}