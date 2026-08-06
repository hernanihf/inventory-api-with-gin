package model

type Page[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

func NewPage[T any](items []T, limit, offset, total int) Page[T] {
	return Page[T]{Items: items, Limit: limit, Offset: offset, Total: total}
}
