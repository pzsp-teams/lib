package cacher

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

type CacheHandler struct {
	Cacher Cacher
	Runner util.TaskRunner
}

func (h *CacheHandler) OnError(err error) {
	h.Runner.Run(func() {
		if shouldClearCache(err) {
			_ = h.Cacher.Clear()
		}
	})
}

func WithErrorClear[T any](
	fn func() (T, error), cacheHandler *CacheHandler,
) (T, error) {
	res, err := fn()
	if err != nil {
		cacheHandler.OnError(err)
		var zero T
		return zero, err
	}
	return res, nil
}

func shouldClearCache(err error) bool {
	if sc, ok := sender.StatusCode(err); ok {
		return sc == http.StatusBadRequest || sc == http.StatusNotFound
	}
	return false
}

func NewCacheHandler(cfg *config.CacheConfig) *CacheHandler {
	if cfg.Mode == config.CacheDisabled {
		return nil
	}

	if cfg.Path == nil {
		defaultPath := defaultCachePath()
		cfg.Path = &defaultPath
	}

	var runner util.TaskRunner = &util.SyncRunner{}
	if cfg.Mode == config.CacheAsync {
		runner = &util.AsyncRunner{}
	}

	var cacher Cacher
	if cfg.Provider == config.CacheProviderJSONFile {
		cacher = newJSONFileCacher(*cfg.Path)
	}

	return &CacheHandler{
		Cacher: cacher,
		Runner: runner,
	}
}

func defaultCachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "pzsp-teams-cache.json"
		}
		return filepath.Join(home, ".pzsp-teams-cache.json")
	}
	p := filepath.Join(dir, "pzsp-teams", "cache.json")
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	return p
}
