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
	if !strings.Contains(out, "stock bajo") || !strings.Contains(out, "coca") {
		t.Errorf("esperaba un warning de stock bajo para coca, log: %s", out)
	}
	if strings.Contains(out, "sprite") {
		t.Errorf("no esperaba ningún log para sprite (stock alto), log: %s", out)
	}
}

func TestCheckLowStock_NoLowStock(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	lister := &fakeProductLister{products: []model.Product{{Name: "sprite", Stock: 10}}}

	checkLowStock(context.Background(), lister)

	if strings.Contains(buf.String(), "stock bajo") {
		t.Errorf("no esperaba ningún warning, log: %s", buf.String())
	}
}

func TestCheckLowStock_StoreError(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	lister := &fakeProductLister{err: errors.New("boom")}

	checkLowStock(context.Background(), lister)

	if !strings.Contains(buf.String(), "no se pudo leer productos") {
		t.Errorf("esperaba un log de error, log: %s", buf.String())
	}
}
