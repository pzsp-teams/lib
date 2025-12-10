package util

func MapSlices[T any, U any](items []T, mapper func(T) U) []U {
	mapped := make([]U, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapper(item))
	}
	return mapped
}