package worker

import (
	"context"
	"inventory_api/model"
	"inventory_api/util"
	"log/slog"
	"time"
)

type productLister interface {
	GetAllProducts(ctx context.Context) ([]model.Product, error)
}

func StartLowStockMonitor(ctx context.Context, s productLister) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("low stock monitor: shutting down")
			return
		case <-ticker.C:
			checkLowStock(ctx, s)
		}
	}
}

func checkLowStock(ctx context.Context, s productLister) {
	const lowStock = 5
	products, err := s.GetAllProducts(ctx)
	if err != nil {
		slog.Error("low stock monitor: could not read products", "error", err)
		return
	}

	low := util.SliceFilter(products, func(p model.Product) bool {
		return p.Stock < lowStock
	})
	for _, p := range low {
		slog.Warn("low stock", "product", p.Name, "stock", p.Stock)
	}
}
