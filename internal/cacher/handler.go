package cacher

import (
	"os"
	"path/filepath"

	"github.com/pzsp-teams/lib/config"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

type CacheHandler struct {
	Cacher Cacher
	Runner util.TaskRunner
}

func (h *CacheHandler) OnError(err *snd.RequestError) {
	h.Runner.Run(func() {
		if shouldClearCache(err) {
			_ = h.Cacher.Clear()
		}
	})
}

func WithErrorClear[T any](
	fn func() (T, *snd.RequestError), cacheHandler *CacheHandler,
) (T, *snd.RequestError) {
	res, err := fn()
	if err != nil {
		cacheHandler.OnError(err)
		var zero T
		return zero, err
	}
	return res, nil
}

func shouldClearCache(err *snd.RequestError) bool {
	if err == nil {
		return false
	}
	switch err.Code {
	case 400, 404, 409, 412, 413, 422:
		return true
	default:
		return false
	}
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
