package util

import (
	"inventory_api/model"
	"slices"
	"testing"
)

func TestSliceFilter_Ints(t *testing.T) {
	got := SliceFilter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	want := []int{2, 4, 6}
	if !slices.Equal(got, want) {
		t.Errorf("esperaba %v, obtuve %v", want, got)
	}
}

func TestSliceFilter_Strings(t *testing.T) {
	got := SliceFilter([]string{"go", "rust", "c", "python"}, func(s string) bool { return len(s) > 2 })
	want := []string{"rust", "python"}
	if !slices.Equal(got, want) {
		t.Errorf("esperaba %v, obtuve %v", want, got)
	}
}

func TestSliceFilter_Products(t *testing.T) {
	products := []model.Product{
		{Name: "coca", Stock: 2},
		{Name: "sprite", Stock: 10},
		{Name: "lays", Stock: 1},
	}
	got := SliceFilter(products, func(p model.Product) bool { return p.Stock < 5 })
	if len(got) != 2 {
		t.Errorf("esperaba 2 productos con poco stock, obtuve %d", len(got))
	}
}

func TestSliceFilter_NoMatches(t *testing.T) {
	got := SliceFilter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
	if len(got) != 0 {
		t.Errorf("esperaba slice vacío, obtuve %v", got)
	}
}

func TestSliceFilter_EmptyInput(t *testing.T) {
	got := SliceFilter([]int{}, func(int) bool { return true })
	if len(got) != 0 {
		t.Errorf("esperaba slice vacío, obtuve %v", got)
	}
}
