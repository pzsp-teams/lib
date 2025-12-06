package util

func Deref[T any](s *T) T {
	var defaultValue T
	if s == nil {
		return defaultValue
	}
	return *s
}
