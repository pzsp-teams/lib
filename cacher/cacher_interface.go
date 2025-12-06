package cacher

type Cacher interface {
	Get(key string) any
	Set(key string, value any)
	Invalidate(key string)
	Clear()
}
