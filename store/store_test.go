package store

import (
	"context"
	"database/sql"
	"errors"
	"inventory_api/model"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "user=myuser dbname=mydb_test password=mypass host=localhost port=5432 sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	testDB = db
	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func resetProducts(t *testing.T) *Store {
	t.Helper()
	if _, err := testDB.Exec(`TRUNCATE TABLE products`); err != nil {
		t.Fatalf("no se pudo limpiar products: %v", err)
	}
	return InitStore(testDB)
}

func TestStore_AddProduct_Upsert(t *testing.T) {
	s := resetProducts(t)
	p := model.Product{Name: "Test", Price: 10.99, Stock: 5}
	if _, err := s.AddProduct(context.Background(), p); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	updated := model.Product{Name: "Test", Price: 12.00, Stock: 8}
	got, err := s.AddProduct(context.Background(), updated)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got != updated {
		t.Errorf("esperaba %v tras el upsert, obtuve %v", updated, got)
	}
}

func TestStore_SearchProduct(t *testing.T) {
	s := resetProducts(t)
	p := model.Product{Name: "Test1", Price: 33.45, Stock: 2}
	s.AddProduct(context.Background(), p)

	found, exists := s.SearchProduct(context.Background(), "Test1")
	if !exists || found != p {
		t.Errorf("esperaba %v, obtuve %v (exists=%v)", p, found, exists)
	}

	if _, exists := s.SearchProduct(context.Background(), "no-existe"); exists {
		t.Errorf("no debería existir")
	}
}

func TestStore_SellProduct(t *testing.T) {
	s := resetProducts(t)
	s.AddProduct(context.Background(), model.Product{Name: "Test", Price: 10.99, Stock: 100})

	updated, err := s.SellProduct(context.Background(), "Test", 50)
	if err != nil || updated.Stock != 50 {
		t.Errorf("esperaba stock 50 sin error, obtuve stock=%d err=%v", updated.Stock, err)
	}

	if _, err := s.SellProduct(context.Background(), "no-existe", 1); !errors.As(err, &model.ProductNotFoundError{}) {
		t.Errorf("esperaba ProductNotFoundError, obtuve %v", err)
	}

	if _, err := s.SellProduct(context.Background(), "Test", 1000); !errors.As(err, &model.InsufficientStockError{}) {
		t.Errorf("esperaba InsufficientStockError, obtuve %v", err)
	}

	if _, err := s.SellProduct(context.Background(), "Test", -10); !errors.As(err, &model.BadFieldInputError{}) {
		t.Errorf("esperaba BadFieldInputError, obtuve %v", err)
	}
}

func TestStore_AddStock(t *testing.T) {
	s := resetProducts(t)
	s.AddProduct(context.Background(), model.Product{Name: "Test", Price: 10.99, Stock: 1})

	updated, err := s.AddStock(context.Background(), "Test", 100)
	if err != nil || updated.Stock != 101 {
		t.Errorf("esperaba stock 101 sin error, obtuve stock=%d err=%v", updated.Stock, err)
	}

	if _, err := s.AddStock(context.Background(), "no-existe", 10); !errors.As(err, &model.ProductNotFoundError{}) {
		t.Errorf("esperaba ProductNotFoundError, obtuve %v", err)
	}
}

func TestStore_GetProducts(t *testing.T) {
	s := resetProducts(t)
	p1 := model.Product{Name: "Test", Price: 10.99, Stock: 1}
	p2 := model.Product{Name: "Test1", Price: 33.45, Stock: 2}
	s.AddProduct(context.Background(), p1)
	s.AddProduct(context.Background(), p2)

	products, total, err := s.GetProducts(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(products) != 2 || total != 2 {
		t.Fatalf("esperaba 2 productos, obtuve %d", len(products))
	}

	got := make(map[string]model.Product, len(products))
	for _, p := range products {
		got[p.Name] = p
	}
	for _, want := range []model.Product{p1, p2} {
		if got[want.Name] != want {
			t.Errorf("esperaba %v, obtuve %v", want, got[want.Name])
		}
	}
}

func TestStore_GetProducts_Empty(t *testing.T) {
	s := resetProducts(t)

	products, total, err := s.GetProducts(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(products) != 0 || total != 0 {
		t.Errorf("esperaba 0 productos, obtuve %d", len(products))
	}
}
