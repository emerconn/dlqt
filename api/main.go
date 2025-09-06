package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux" // HTTP router for handling different endpoints
)

// main is the entry point of the API service
// It sets up the HTTP server and routing for the API service
func main() {
	// Service startup logging with timestamp for debugging deployment issues
	log.Println("=== Starting DLQT API Service ===")
	log.Printf("Timestamp: %s", time.Now().Format(time.RFC3339))

	// Create an instance of our API service - this handles the business logic
	apiService := &APIService{}
	log.Println("API service instance created")

	// Setup HTTP server using Gorilla Mux router
	// Mux provides more advanced routing features than the standard library
	log.Println("Setting up HTTP router...")
	r := mux.NewRouter()

	// Register trigger endpoint - requires dlq.retrigger scope
	log.Println("Registering /trigger endpoint...")
	r.HandleFunc("/trigger", apiService.handleTrigger).Methods("POST")

	// Register fetch endpoint - requires dlq.fetch scope
	log.Println("Registering /fetch endpoint...")
	r.HandleFunc("/fetch", apiService.handleFetch).Methods("GET")

	// Log the service configuration for operational visibility
	log.Println("=== API service configuration complete ===")
	log.Println("Available endpoints:")
	log.Println("  POST /trigger - Trigger operations (requires dlq.retrigger scope)")
	log.Println("  GET /fetch - Fetch operations (requires dlq.fetch scope)")

	// Start the HTTP server on port 8080
	// log.Fatal will terminate the program if the server fails to start
	log.Println("Starting API service on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
