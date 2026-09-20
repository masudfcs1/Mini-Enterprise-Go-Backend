package response

import (
	"encoding/json"
	"net/http"
)

// StandardResponse is the unified envelope for all API responses.
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// JSON sends a custom payload with a specific HTTP status code.
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Success sends a standard success response envelope.
func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends a standard error response envelope.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, StandardResponse{
		Success: false,
		Error:   message,
	})
}

// ErrorWithDetails sends a standard error response with additional error details.
func ErrorWithDetails(w http.ResponseWriter, status int, message string, details interface{}) {
	JSON(w, status, StandardResponse{
		Success: false,
		Message: message,
		Error:   details,
	})
}
