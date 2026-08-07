package worker

import (
	"bytes"
	"context"
	"errors"
	"inventory_api/model"
	"log/slog"
	"strings"
	"testing"
)

type fakeProductLister struct {
	products []model.Product
	err      error
}

func (f *fakeProductLister) GetAllProducts(_ context.Context) ([]model.Product, error) {
	return f.products, f.err
}

func TestCheckLowStock_WarnsOnLowStock(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	lister := &fakeProductLister{products: []model.Product{
		{Name: "coca", Stock: 2},
		{Name: "sprite", Stock: 10},
	}}

	checkLowStock(context.Background(), lister)

	out := buf.String()
	if !strings.Contains(out, "low stock") || !strings.Contains(out, "coca") {
		t.Errorf("expects low stock warning for coca, log: %s", out)
	}
	if strings.Contains(out, "sprite") {
		t.Errorf("not expect logs for sprite (high stock), log: %s", out)
	}
}

func TestCheckLowStock_NoLowStock(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	lister := &fakeProductLister{products: []model.Product{{Name: "sprite", Stock: 10}}}

	checkLowStock(context.Background(), lister)

	if strings.Contains(buf.String(), "stock bajo") {
		t.Errorf("not expect warnings, log: %s", buf.String())
	}
}

func TestCheckLowStock_StoreError(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	lister := &fakeProductLister{err: errors.New("boom")}

	checkLowStock(context.Background(), lister)

	if !strings.Contains(buf.String(), "could not read products") {
		t.Errorf("expects error log, log: %s", buf.String())
	}
}
