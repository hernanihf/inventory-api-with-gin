package handler

import (
	"errors"
	"fmt"
	"inventory_api/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, model.ErrorResponse{Error: message})
}

func respondBindError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make([]string, 0, len(ve))
		for _, fe := range ve {
			details = append(details, fmt.Sprintf("%s: rule failed '%s'", fe.Field(), fe.Tag()))
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "validation failed", Details: details})
		return
	}
	respondError(c, http.StatusBadRequest, "invalid request body")
}
