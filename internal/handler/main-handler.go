package handler

import (
	"encoding/json"
	"net/http"

	"change-type-image.com/internal/models"
)

func NoRoutes(writer http.ResponseWriter, request *http.Request) {
	res := models.SuccessResponse{Status: http.StatusAccepted, Message: "Routes not found", Success: true}

	js, err := json.Marshal(res)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(http.StatusNotFound)
	writer.Write(js)
}
