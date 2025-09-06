package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux" // HTTP router for handling different endpoints
)

// Helper function for finding the minimum of two integers
// Used when safely truncating strings to avoid index out of bounds errors
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// main is the entry point of the authentication service
// It sets up the HTTP server, middleware, and routing for the auth service
func main() {
	// Service startup logging with timestamp for debugging deployment issues
	log.Println("=== Starting DLQT Auth Service ===")
	log.Printf("Timestamp: %s", time.Now().Format(time.RFC3339))

	// Create an instance of our auth service - this handles the business logic
	authService := &AuthService{}
	log.Println("Auth service instance created")

	// Setup HTTP server using Gorilla Mux router
	// Mux provides more advanced routing features than the standard library
	log.Println("Setting up HTTP router...")
	r := mux.NewRouter()

	// Add authentication middleware to ALL routes
	// This middleware runs before any handler and validates JWT tokens
	log.Println("Adding auth middleware...")
	r.Use(authMiddleware)

	// Register our main endpoint for checking authorization
	// Only POST requests to /check-auth will be handled by this route
	log.Println("Registering /check-auth endpoint...")
	r.HandleFunc("/check-auth", authService.handleCheckAuth).Methods("POST")

	// Log the service configuration for operational visibility
	log.Println("=== Auth service configuration complete ===")
	log.Println("Available endpoints:")
	log.Println("  POST /check-auth - Check user authorization for namespace/queue")

	// Start the HTTP server on port 8080
	// log.Fatal will terminate the program if the server fails to start
	log.Println("Starting auth service on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
