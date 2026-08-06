package model

import "fmt"

type ProductNotFoundError struct {
	ProductName string
}

func (e ProductNotFoundError) Error() string {
	return fmt.Sprintf("Product %s not found", e.ProductName)
}

type InsufficientStockError struct {
	ProductName string
}

func (e InsufficientStockError) Error() string {
	return fmt.Sprintf("Product %s has insufficient stock", e.ProductName)
}

type BadFieldInputError struct {
	Description string
}

func (e BadFieldInputError) Error() string {
	return fmt.Sprintf("Input error: %s", e.Description)
}
