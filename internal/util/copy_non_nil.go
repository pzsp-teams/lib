// Package util contains shared helpers used across the project, including:
//   - pointer and slice utilities (Deref, MapSlices, CopyNonNil),
//   - hashing helpers (HashWithPepper),
//   - simple validators for IDs/emails,
//   - simple sync/async task runners.
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
