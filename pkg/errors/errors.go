package errors

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-mini-setup/internal/database/db"
	"go-mini-setup/pkg/response"
)

// AppError represents an application-level domain error.
type AppError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: message,
	}
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Code:    "CONFLICT",
		Message: message,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Status:  http.StatusUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

func NewInternalError(message string) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: message,
	}
}

// HandleError standardizes error responses across HTTP handlers.
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		response.Error(w, appErr.Status, appErr.Message)
		return
	}

	// Prisma specific error checks
	if db.IsErrNotFound(err) {
		response.Error(w, http.StatusNotFound, "Resource not found")
		return
	}

	errStr := err.Error()
	if strings.Contains(errStr, "Unique constraint") || strings.Contains(errStr, "unique constraint") {
		response.Error(w, http.StatusConflict, "A record with this unique identifier already exists")
		return
	}

	log.Printf("[Error Handler] Unhandled error: %v", err)
	response.Error(w, http.StatusInternalServerError, "Internal server error")
}
