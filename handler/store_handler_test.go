package handler

import (
	"encoding/json"
	"inventory_api/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStoreHandler_GetProducts_Found(t *testing.T) {
	gin.SetMode(gin.TestMode)
	coca := model.Product{Name: "coca", Price: 2.5, Stock: 10}
	pepsi := model.Product{Name: "pepsi", Price: 2, Stock: 12}
	sprite := model.Product{Name: "sprite", Price: 2.0, Stock: 8}
	store := newFakeStore(coca, pepsi, sprite)
	handler := NewStore(store)

	router := gin.New()
	router.GET("/products", handler.GetProducts)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	var response model.Page[model.Product]
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	got := make(map[string]model.Product, len(response.Items))
	for _, p := range response.Items {
		got[p.Name] = p
	}
	for _, want := range []model.Product{coca, pepsi, sprite} {
		if got[want.Name] != want {
			t.Errorf("esperaba %v, obtuve %v", want, got[want.Name])
		}
	}
}

func TestStoreHandler_GetProducts_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore())

	router := gin.New()
	router.GET("/products", handler.GetProducts)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	var response model.Page[model.Product]
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	if len(response.Items) != 0 {
		t.Errorf("esperaba 0 elementos, obtuve %d", len(response.Items))
	}
}

func TestStoreHandler_GetProducts_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testCases := map[string]string{
		"no numérico":      "abc",
		"cero":             "0",
		"negativo":         "-5",
		"excede el máximo": "101",
	}
	for name, limit := range testCases {
		t.Run(name, func(t *testing.T) {
			handler := NewStore(newFakeStore())
			router := gin.New()
			router.GET("/products", handler.GetProducts)

			req := httptest.NewRequest(http.MethodGet, "/products?limit="+limit, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("esperaba 400 para limit=%q, obtuve %d", limit, rec.Code)
			}
		})
	}
}

func TestStoreHandler_GetProducts_InvalidOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testCases := map[string]string{
		"no numérico": "abc",
		"negativo":    "-1",
	}
	for name, offset := range testCases {
		t.Run(name, func(t *testing.T) {
			handler := NewStore(newFakeStore())
			router := gin.New()
			router.GET("/products", handler.GetProducts)

			req := httptest.NewRequest(http.MethodGet, "/products?offset="+offset, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("esperaba 400 para offset=%q, obtuve %d", offset, rec.Code)
			}
		})
	}
}

func TestStoreHandler_GetProducts_CustomPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore(
		model.Product{Name: "a"},
		model.Product{Name: "b"},
		model.Product{Name: "c"},
	))
	router := gin.New()
	router.GET("/products", handler.GetProducts)

	req := httptest.NewRequest(http.MethodGet, "/products?limit=1&offset=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	var page model.Page[model.Product]
	if err := json.NewDecoder(rec.Body).Decode(&page); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	if page.Limit != 1 || page.Offset != 1 || page.Total != 3 {
		t.Errorf("metadata incorrecta: %+v", page)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "b" {
		t.Errorf("esperaba solo 'b', obtuve %+v", page.Items)
	}
}

func TestStoreHandler_GetProduct_Found(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newFakeStore(model.Product{Name: "coca", Price: 2.5, Stock: 10})
	handler := NewStore(store)

	router := gin.New()
	router.GET("/products/:name", handler.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/products/coca", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	var got model.Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	if got.Name != "coca" {
		t.Errorf("esperaba coca, obtuve %s", got.Name)
	}
}

func TestStoreHandler_GetProduct_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore())

	router := gin.New()
	router.GET("/products/:name", handler.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/products/nope", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba 404, obtuve %d", rec.Code)
	}
}

func TestStoreHandler_SellProduct_InsufficientStock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore(model.Product{Name: "coca", Price: 2.5, Stock: 1}))

	router := gin.New()
	router.PUT("/products/:name/sell", handler.SellProduct)

	req := httptest.NewRequest(http.MethodPut, "/products/coca/sell", strings.NewReader(`{"quantity": 5}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuve %d", rec.Code)
	}
}

func TestStoreHandler_SellProduct_ProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore(model.Product{Name: "coca", Price: 2.5, Stock: 1}))

	router := gin.New()
	router.PUT("/products/:name/sell", handler.SellProduct)

	req := httptest.NewRequest(http.MethodPut, "/products/cocoa/sell", strings.NewReader(`{"quantity": 5}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba 404, obtuve %d", rec.Code)
	}
}

func TestStoreHandler_SellProduct_NegativeInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore(model.Product{Name: "coca", Price: 2.5, Stock: 1}))

	router := gin.New()
	router.PUT("/products/:name/sell", handler.SellProduct)

	req := httptest.NewRequest(http.MethodPut, "/products/coca/sell", strings.NewReader(`{"quantity": -5}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuve %d", rec.Code)
	}
}

func TestStoreHandler_SellProduct_Ok(t *testing.T) {
	gin.SetMode(gin.TestMode)
	coca := model.Product{Name: "coca", Price: 2.5, Stock: 1}
	handler := NewStore(newFakeStore(coca))

	router := gin.New()
	router.PUT("/products/:name/sell", handler.SellProduct)

	req := httptest.NewRequest(http.MethodPut, "/products/coca/sell", strings.NewReader(`{"quantity": 1}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba 200, obtuve %d", rec.Code)
	}
	var response model.Product
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	if response.Stock != 0 {
		t.Errorf("esperaba %d stock de coca, obtuve %d", coca.Stock-1, response.Stock)
	}
}

func TestStoreHandler_AddProductStock_Ok(t *testing.T) {
	gin.SetMode(gin.TestMode)
	coca := model.Product{Name: "coca", Price: 2.5, Stock: 1}
	handler := NewStore(newFakeStore(coca))

	router := gin.New()
	router.PUT("/products/:name/addStock", handler.AddProductStock)

	req := httptest.NewRequest(http.MethodPut, "/products/coca/addStock", strings.NewReader(`{"quantity": 1}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba 200, obtuve %d", rec.Code)
	}
	var response model.Product
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	if response.Stock != 2 {
		t.Errorf("esperaba %d stock de coca, obtuve %d", coca.Stock+1, response.Stock)
	}
}

func TestStoreHandler_AddProductStock_ProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore(model.Product{Name: "coca", Price: 2.5, Stock: 1}))

	router := gin.New()
	router.PUT("/products/:name/addStock", handler.AddProductStock)

	req := httptest.NewRequest(http.MethodPut, "/products/cocoa/addStock", strings.NewReader(`{"quantity": 5}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba 404, obtuve %d", rec.Code)
	}
}

func TestStoreHandler_AddProductStock_NegativeInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore(model.Product{Name: "coca", Price: 2.5, Stock: 1}))

	router := gin.New()
	router.PUT("/products/:name/addStock", handler.AddProductStock)

	req := httptest.NewRequest(http.MethodPut, "/products/coca/addStock", strings.NewReader(`{"quantity": -5}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuve %d", rec.Code)
	}
}

func TestStoreHandler_AddProduct_Ok(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStore(newFakeStore())

	router := gin.New()
	router.POST("/products", handler.AddProduct)

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{"name": "testing","price": 1.0, "quantity": 1}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("esperaba 201, obtuve %d", rec.Code)
	}
	var response model.Product
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v", err)
	}
	if response.Name != "testing" {
		t.Errorf("esperaba %s obtuve %s", response.Name, response.Name)
	}
}
