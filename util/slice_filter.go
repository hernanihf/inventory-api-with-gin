package util

func SliceFilter[T any](items []T, predicate func(T) bool) []T {
	var out []T
	for _, item := range items {
		if predicate(item) {
			out = append(out, item)
		}
	}
	return out
}
