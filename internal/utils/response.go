package utils

import (
	"encoding/json"
	"net/http"

	"change-type-image.com/internal/models"
)

func ErrorResponse(w http.ResponseWriter, status int, message string, errorDetail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Status:      status,
		Message:     message,
		ErrorDetail: errorDetail,
	})
}

func SuccessImageResponse(w http.ResponseWriter, success bool, status int, message string, data any) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Status:  status,
		Message: message,
		Success: success,
		Data:    data,
	})
}
