package cacher

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONFileCacher struct {
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
	err := json.Unmarshal(cacher.cache[key], &slice)
	if err != nil {
		slice = []string{}
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
	cacher.cache = make(map[string]json.RawMessage)
	data, err := json.MarshalIndent(cacher.cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacher.file, data, 0o644)
}
