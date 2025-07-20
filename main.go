package main

import (
	"log"
	"net/http"
	"os"

	"smit/server/api/handlers"
	"smit/server/api/storage"
)

func main() {
	// Get configuration from environment
	port := getEnv("SERVER_PORT", "1234")
	dataFilePath := getEnv("DATA_FILE_PATH", "./data/data.json")

	// Setup and start server
	server, err := setupServer(dataFilePath)
	if err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}

	// Start server
	log.Printf("Starting SMIT Network API server on port %s", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupServer sets up the HTTP server with all routes and middleware
func setupServer(dataFilePath string) (http.Handler, error) {
	// Initialize storage
	store, err := storage.NewJSONStorage(dataFilePath)
	if err != nil {
		return nil, err
	}

	// Initialize handlers
	handler := handlers.NewHandler(store)

	// Setup routes
	mux := http.NewServeMux()

	// VLAN endpoints
	mux.HandleFunc("/api/v1/vlans", handler.VLANHandler)
	mux.HandleFunc("/api/v1/vlans/", handler.VLANHandler)

	// Health endpoint
	mux.HandleFunc("/health", handler.HealthCheck)

	// Add CORS middleware
	return cors(mux), nil
}

// cors adds CORS headers to responses
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
