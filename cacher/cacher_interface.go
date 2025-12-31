// Package cacher contains caching utilities for the library, including:
//   - the Cacher interface,
//   - a JSON file-backed cacher,
//   - key builders for teams, channels, chats and members.
package cacher

type Cacher interface {
	Get(key string) (value any, found bool, err error)
	Set(key string, value any) error
	Invalidate(key string) error
	Clear() error
}
