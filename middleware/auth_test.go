package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuth_ValidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	apiKey := "secret"

	router := gin.New()
	router.Use(Auth(apiKey))
	router.POST("/products", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/products", nil)
	req.Header.Set("X-API-Key", apiKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if !called {
		t.Errorf("next handler call was expected to be called")
	}
}

func TestAuth_MissingKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	apiKey := "secret"
	router := gin.New()

	router.Use(Auth(apiKey))
	router.POST("/products", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/products", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if called {
		t.Errorf("next handler was not expected to be called")
	}
}

func TestAuth_WrongKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	apiKey := "secret"

	router := gin.New()
	router.Use(Auth(apiKey))
	router.POST("/products", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/products", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if called {
		t.Errorf("next handler was not expected to be called")
	}
}
