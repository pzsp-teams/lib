package util

func CopyNonNil[T any](items []*T) []T {
	local := make([]T, 0, len(items))
	for _, item := range items {
		if item != nil {
			local = append(local, *item)
		}
	}
	return local
}