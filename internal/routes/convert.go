package routes

import (
	"net/http"

	"change-type-image.com/internal/handler"
)

func ConvertRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handler.ConvertImages(w, r)
	} else {
		handler.NoRoutes(w, r)
	}
}
