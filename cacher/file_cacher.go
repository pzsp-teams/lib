package cacher

import (
	"encoding/json"
	"fmt"
	"os"
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

func (cacher *JSONFileCacher) Get(key string) (value any, found bool, err error) {
	cacher.mu.Lock()
	defer cacher.mu.Unlock()
	if cacher.loaded {
		return cacher.getFromCache(key)
	}
	if err := cacher.loadCache(); err != nil {
		return nil, false, err
	}
	return cacher.getFromCache(key)
}

func (cacher *JSONFileCacher) getFromCache(key string) (value any, found bool, err error) {
	data, ok := cacher.cache[key]
	if !ok {
		return nil, false, nil
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false, err
	}
	return result, true, nil
}

func (cacher *JSONFileCacher) loadCache() error {
	data, err := os.ReadFile(cacher.file)
	if err != nil {
		if os.IsNotExist(err) {
			cacher.loaded = true
			cacher.cache = make(map[string]json.RawMessage)
			return nil
		} else {
			return err
		}
	}
	if len(data) == 0 {
		cacher.loaded = true
		cacher.cache = make(map[string]json.RawMessage)
	} else {
		if err := json.Unmarshal(data, &cacher.cache); err != nil {
			return err
		}
		cacher.loaded = true
	}
	return nil
}

func (cacher *JSONFileCacher) Set(key string, value any) error {
	cacher.mu.Lock()
	defer cacher.mu.Unlock()
	if !cacher.loaded {
		if err := cacher.loadCache(); err != nil {
			return err
		}
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("JSONFileCacher.Set: expected string, got %T", value)
	}
	var slice []string
	if raw, exists := cacher.cache[key]; exists && raw != nil {
		if err := json.Unmarshal(raw, &slice); err != nil {
			slice = nil
		}
	}
	seen := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		seen[s] = struct{}{}
	}
	if _, exists := seen[str]; exists {
		return nil
	}
	slice = append(slice, str)
	record, err := json.Marshal(slice)
	if err != nil {
		return err
	}
	cacher.cache[key] = record
	data, err := json.MarshalIndent(cacher.cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacher.file, data, 0o644)
}

func (cacher *JSONFileCacher) Invalidate(key string) error {
	cacher.mu.Lock()
	defer cacher.mu.Unlock()
	if !cacher.loaded {
		err := cacher.loadCache()
		if err != nil {
			return err
		}
	}
	delete(cacher.cache, key)
	data, err := json.MarshalIndent(cacher.cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacher.file, data, 0o644)
}

func (cacher *JSONFileCacher) Clear() error {
	cacher.mu.Lock()
	defer cacher.mu.Unlock()
	cacher.cache = make(map[string]json.RawMessage)
	cacher.loaded = true
	data, err := json.MarshalIndent(cacher.cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacher.file, data, 0o644)
}
