package cacher

import (
	"os"
	"path/filepath"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/util"
)

type CacheHandler struct {
	Cacher Cacher
	Runner util.TaskRunner
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

func (c *CacheHandler) OnError() {
	c.Runner.Run(func() {
		_ = c.Cacher.Clear()
	})
}

func WithErrorClear[T any](
	c *CacheHandler,
	fn func() (T, error),
) (T, error) {
	res, err := fn()
	if err != nil {
		c.OnError()
		var zero T
		return zero, err
	}
	return res, err
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
