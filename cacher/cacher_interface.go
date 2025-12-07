package cacher

type Cacher interface {
	Get(key string) (value any, found bool, err error)
	Set(key string, value any) error
	Invalidate(key string) error
	Clear() error
}
