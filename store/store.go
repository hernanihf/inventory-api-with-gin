package store

import (
	"context"
	"database/sql"
	"errors"
	"inventory_api/model"
)

type Store struct {
	db *sql.DB
}

func InitStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) AddProduct(ctx context.Context, p model.Product) (model.Product, error) {
	var updated model.Product
	err := s.db.QueryRowContext(
		ctx, `INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)
         ON CONFLICT (name) DO UPDATE SET price = $2, stock = $3
         RETURNING name, price, stock`,
		p.Name, p.Price, p.Stock,
	).Scan(&updated.Name, &updated.Price, &updated.Stock)
	return updated, err
}

func (s *Store) GetProducts(ctx context.Context, limit, offset int) ([]model.Product, int, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT name, price, stock FROM products ORDER BY name LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.Name, &p.Price, &p.Stock); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&total); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (s *Store) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, price, stock FROM products ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.Name, &p.Price, &p.Stock); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *Store) SearchProduct(ctx context.Context, name string) (model.Product, bool) {
	var p model.Product
	err := s.db.QueryRowContext(ctx,
		`SELECT name, price, stock FROM products WHERE name = $1`, name,
	).Scan(&p.Name, &p.Price, &p.Stock)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, false
	}
	return p, err == nil
}

func (s *Store) SellProduct(ctx context.Context, name string, quantity int) (model.Product, error) {
	if quantity < 0 {
		return model.Product{}, model.BadFieldInputError{Description: "quantity must be positive"}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback()

	var stock int
	err = tx.QueryRowContext(ctx, `SELECT stock FROM products WHERE name = $1 FOR UPDATE`, name).Scan(&stock)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, model.ProductNotFoundError{ProductName: name}
	}
	if err != nil {
		return model.Product{}, err
	}
	if stock < quantity {
		return model.Product{}, model.InsufficientStockError{ProductName: name}
	}

	var updated model.Product
	err = tx.QueryRowContext(ctx,
		`UPDATE products SET stock = stock - $1 WHERE name = $2
         RETURNING name, price, stock`,
		quantity, name,
	).Scan(&updated.Name, &updated.Price, &updated.Stock)
	if err != nil {
		return model.Product{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Product{}, err
	}
	return updated, nil
}

func (s *Store) AddStock(ctx context.Context, name string, quantity int) (model.Product, error) {
	if quantity < 0 {
		return model.Product{}, model.BadFieldInputError{Description: "quantity must be positive"}
	}

	var updated model.Product
	err := s.db.QueryRowContext(
		ctx, `UPDATE products SET stock = stock + $1 WHERE name = $2
         RETURNING name, price, stock`,
		quantity, name,
	).Scan(&updated.Name, &updated.Price, &updated.Stock)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, model.ProductNotFoundError{ProductName: name}
	}
	if err != nil {
		return model.Product{}, err
	}
	return updated, nil
}
