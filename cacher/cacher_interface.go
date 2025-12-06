package cacher

type Cacher interface {
	Get(key string) (any, bool, error)
	Set(key string, value any) error
	Invalidate(key string) error
	Clear() error
}
