package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"change-type-image.com/internal/routes"
	"change-type-image.com/internal/utils"
	"golang.org/x/net/http2"
)

func main() {

	// ! LOAD THE ENV FILE
	utils.LoadEnv(".env")

	// ! SETUP THE SERVER
	server, port := setupServer()

	http2.ConfigureServer(server, &http2.Server{})
	fmt.Println("Server is running on port:", port)

	cert := os.Getenv("CERT_PEM")
	key := os.Getenv("KEY_PEM")

	if cert == "" || key == "" {
		log.Fatal("CERT_PEM and KEY_PEM must be set")
	}

	// ! PROD
	// err := server.ListenAndServeTLS(cert, key)
	err := server.ListenAndServe()

	if err != nil {
		log.Fatalln("Could not server", err)
	}

}

// ! SETUP SERVER
func setupServer() (*http.Server, string) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	routes.InitRoutes(mux)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           enableCORS(mux),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return server, port
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5500")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
