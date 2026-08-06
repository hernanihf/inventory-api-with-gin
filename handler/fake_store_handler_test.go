package handler

import (
	"context"
	"inventory_api/model"
	"sort"
)

type fakeStore struct {
	products map[string]model.Product
}

func newFakeStore(products ...model.Product) *fakeStore {
	m := make(map[string]model.Product)
	for _, p := range products {
		m[p.Name] = p
	}
	return &fakeStore{products: m}
}

func (f *fakeStore) AddProduct(_ context.Context, p model.Product) (model.Product, error) {
	f.products[p.Name] = p
	return p, nil
}

func (f *fakeStore) GetProducts(_ context.Context, limit, offset int) ([]model.Product, int, error) {
	var out []model.Product
	for _, p := range f.products {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	total := len(out)
	if offset > len(out) {
		return []model.Product{}, total, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

func (f *fakeStore) GetAllProducts(_ context.Context) ([]model.Product, error) {
	products, _, err := f.GetProducts(context.Background(), len(f.products), 0)
	return products, err
}

func (f *fakeStore) SearchProduct(_ context.Context, name string) (model.Product, bool) {
	p, ok := f.products[name]
	return p, ok
}

func (f *fakeStore) SellProduct(_ context.Context, name string, quantity int) (model.Product, error) {
	p, ok := f.products[name]
	if quantity < 0 {
		return model.Product{}, model.BadFieldInputError{Description: name}
	}
	if !ok {
		return model.Product{}, model.ProductNotFoundError{ProductName: name}
	}
	if p.Stock < quantity {
		return model.Product{}, model.InsufficientStockError{ProductName: name}
	}
	p.Stock -= quantity
	f.products[name] = p
	return p, nil
}

func (f *fakeStore) AddStock(_ context.Context, name string, quantity int) (model.Product, error) {
	p, ok := f.products[name]
	if quantity < 0 {
		return model.Product{}, model.BadFieldInputError{Description: name}
	}
	if !ok {
		return model.Product{}, model.ProductNotFoundError{ProductName: name}
	}
	p.Stock += quantity
	f.products[name] = p
	return p, nil
}
