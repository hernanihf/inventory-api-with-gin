package handler

import (
	"context"
	"errors"
	"fmt"
	"inventory_api/model"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type productStore interface {
	AddProduct(ctx context.Context, p model.Product) (model.Product, error)
	GetProducts(ctx context.Context, limit, offset int) ([]model.Product, int, error)
	GetAllProducts(ctx context.Context) ([]model.Product, error)
	SearchProduct(ctx context.Context, name string) (model.Product, bool)
	SellProduct(ctx context.Context, name string, quantity int) (model.Product, error)
	AddStock(ctx context.Context, name string, quantity int) (model.Product, error)
}

type productUpdateDTO struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type StoreHandler struct {
	productStore productStore
}

func NewStore(productStore productStore) *StoreHandler {
	return &StoreHandler{productStore: productStore}
}

func (s *StoreHandler) HelloWorld(c *gin.Context) {
	_, err := fmt.Fprintln(c.Writer, "<h1>WELCOME TO INVENTORY API</h>")
	if err != nil {
		return
	}
}

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

// GetProducts 	godoc
// @Summary      List product
// @Tags         products
// @Produce      json
// @Param        limit   query     int  false  "Quantity od products (default 20, max 100)"
// @Param        offset  query     int  false  "From position (default 0)"
// @Success      200  {object}  model.Page[model.Product]
// @Failure      400  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /products [get]
func (s *StoreHandler) GetProducts(c *gin.Context) {
	limit, offset, err := parsePagination(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	products, total, err2 := s.productStore.GetProducts(c.Request.Context(), limit, offset)
	if err2 != nil {
		slog.Error("failed to get products", "error", err2)
		respondError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, model.NewPage(products, limit, offset, total))
}

func parsePagination(c *gin.Context) (limit, offset int, err error) {
	limit = defaultPageLimit
	if v := c.Query("limit"); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit <= 0 || limit > maxPageLimit {
			return 0, 0, fmt.Errorf("limit must be a number between 1 and %d", maxPageLimit)
		}
	}
	if v := c.Query("offset"); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("offset must be a non-negative number")
		}
	}
	return limit, offset, nil
}

// GetProduct godoc
// @Summary      Search product by name
// @Tags         products
// @Produce      json
// @Param        name  path      string  true  "Product name"
// @Success      200   {object}  model.Product
// @Failure      404   {object}  model.ErrorResponse
// @Router       /products/{name} [get]
func (s *StoreHandler) GetProduct(c *gin.Context) {
	name := c.Param("name")
	product, ok := s.productStore.SearchProduct(c.Request.Context(), name)
	if !ok {
		respondError(c, http.StatusNotFound, fmt.Sprintf("product %q not found", name))
		return
	}
	c.JSON(http.StatusOK, product)
}

// AddProduct godoc
// @Summary      Create or update product
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        product  body      model.Product  true  "Product"
// @Success      201      {object}  model.Product
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /products [post]
func (s *StoreHandler) AddProduct(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		respondBindError(c, err)
		return
	}
	addedProduct, err := s.productStore.AddProduct(c.Request.Context(), product)
	if err != nil {
		slog.Error("failed to add product", "error", err)
		respondError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusCreated, addedProduct)
}

// SellProduct godoc
// @Summary      Sell stock of a product
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        name      path      string         	 true  	"Product name"
// @Param        quantity  body      productUpdateDTO  	true  	"Quantity to sell"
// @Success      200  {object}  model.Product
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /products/{name}/sell [put]
func (s *StoreHandler) SellProduct(c *gin.Context) {
	var productQuantityDTO productUpdateDTO
	if err := c.ShouldBindJSON(&productQuantityDTO); err != nil {
		respondBindError(c, err)
		return
	}
	nameParam := c.Param("name")
	prod, err2 := s.productStore.SellProduct(c.Request.Context(), nameParam, productQuantityDTO.Quantity)
	if err2 != nil {
		stockError := model.InsufficientStockError{}
		if errors.As(err2, &stockError) {
			respondError(c, http.StatusBadRequest, err2.Error())
			return
		}
		foundError := model.ProductNotFoundError{}
		if errors.As(err2, &foundError) {
			respondError(c, http.StatusNotFound, err2.Error())
			return
		}
		badInputError := model.BadFieldInputError{}
		if errors.As(err2, &badInputError) {
			respondError(c, http.StatusBadRequest, err2.Error())
			return
		}
		slog.Error("failed to sell product", "error", err2, "product", nameParam)
		respondError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, prod)
}

// AddProductStock 	godoc
// @Summary      	Add product stock
// @Tags         	products
// @Accept       	json
// @Produce      	json
// @Security     	ApiKeyAuth
// @Param        	name      path      string          	true  "Name of the product"
// @Param        	quantity  body      productUpdateDTO  	true  "Quantity to add"
// @Success      	200  {object}  model.Product
// @Failure      	400  {object}  model.ErrorResponse
// @Failure      	404  {object}  model.ErrorResponse
// @Failure      	500  {object}  model.ErrorResponse
// @Router       	/products/{name}/addStock [put]
func (s *StoreHandler) AddProductStock(c *gin.Context) {
	var productQuantityDTO productUpdateDTO
	if err := c.ShouldBindJSON(&productQuantityDTO); err != nil {
		respondBindError(c, err)
		return
	}
	nameParam := c.Param("name")
	prod, err2 := s.productStore.AddStock(c.Request.Context(), nameParam, productQuantityDTO.Quantity)
	if err2 != nil {
		foundError := model.ProductNotFoundError{}
		if errors.As(err2, &foundError) {
			respondError(c, http.StatusNotFound, err2.Error())
			return
		}
		badInputError := model.BadFieldInputError{}
		if errors.As(err2, &badInputError) {
			respondError(c, http.StatusBadRequest, err2.Error())
			return
		}
		slog.Error("failed to add product stock", "error", err2, "product", nameParam)
		respondError(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, prod)
}
