package cacher

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sync"
)

type JSONFileCacher struct {
	mu     sync.Mutex
	file   string
	cache  map[string]json.RawMessage
	loaded bool
}

func NewJSONFileCacher(path string) Cacher {
	return &JSONFileCacher{
		file: path,
	}
}

func (c *JSONFileCacher) Get(key string) (value any, found bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return c.getFromCache(key)
	}
	if err := c.loadCache(); err != nil {
		return nil, false, err
	}
	return c.getFromCache(key)
}

func (c *JSONFileCacher) getFromCache(key string) (value any, found bool, err error) {
	data, ok := c.cache[key]
	if !ok {
		return nil, false, nil
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false, err
	}
	return result, true, nil
}

func (c *JSONFileCacher) loadCache() error {
	data, err := os.ReadFile(c.file)
	if err != nil {
		if os.IsNotExist(err) {
			c.loaded = true
			c.cache = make(map[string]json.RawMessage)
			return nil
		}
		return err
	}
	if len(data) == 0 {
		c.cache = make(map[string]json.RawMessage)
	} else {
		if err := json.Unmarshal(data, &c.cache); err != nil {
			return err
		}
	}
	c.loaded = true
	return nil
}

func (c *JSONFileCacher) Set(key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureLoadedLocked(); err != nil {
		return err
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("JSONFileCacher.Set: expected string, got %T", value)
	}
	var slice []string
	if raw, exists := c.cache[key]; exists && raw != nil {
		if err := json.Unmarshal(raw, &slice); err != nil {
			slice = nil
		}
	}
	if slices.Contains(slice, str) {
		return nil
	}
	slice = append(slice, str)
	record, err := json.Marshal(slice)
	if err != nil {
		return err
	}
	c.cache[key] = record
	return c.persistLocked()
}

func (c *JSONFileCacher) Invalidate(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureLoadedLocked(); err != nil {
		return err
	}
	delete(c.cache, key)
	return c.persistLocked()
}

func (c *JSONFileCacher) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]json.RawMessage)
	c.loaded = true
	return c.persistLocked()
}

func (c *JSONFileCacher) ensureLoadedLocked() error {
	if c.loaded {
		return nil
	}
	return c.loadCache()
}

func (c *JSONFileCacher) persistLocked() error {
	data, err := json.MarshalIndent(c.cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.file, data, 0o644)
}
