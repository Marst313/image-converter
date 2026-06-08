package routes

import (
	"net/http"
)

func InitRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/convert/image", ConvertRoutes)
}
