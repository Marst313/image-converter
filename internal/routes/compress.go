package routes

import (
	"net/http"

	"change-type-image.com/internal/handler"
)

func CompressRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handler.CompressImage(w, r)

	} else {
		handler.NoRoutes(w, r)
	}
}
