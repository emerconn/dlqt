package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// handleCheckAuth is the main HTTP handler for the /check-auth endpoint
// It receives POST requests with namespace/queue info and returns authorization decisions
// This handler only runs AFTER the authMiddleware has validated the JWT token
func (a *AuthService) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== CHECK-AUTH HANDLER START ===")
	log.Printf("Request from: %s", r.RemoteAddr)

	// === STEP 1: GET USER ID FROM CONTEXT ===
	// The auth middleware already validated the token and extracted the user ID
	// We can safely get it from the request context
	userID := r.Context().Value("userID").(string)
	log.Printf("Authenticated user ID: %s", userID)

	// === STEP 2: PARSE REQUEST BODY ===
	// The client sends JSON with namespace and queue information
	log.Println("Reading request body...")
	var req CheckAuthRequest

	// Decode the JSON request body into our struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return // Stop processing - bad request format
	}

	// Log the parsed request data for debugging
	log.Printf("Request decoded successfully:")
	log.Printf("  Namespace: '%s'", req.Namespace)
	log.Printf("  Queue: '%s'", req.Queue)

	// === STEP 3: CHECK AUTHORIZATION ===
	// This is where the actual business logic happens
	// We check if this user is allowed to access the specified namespace/queue
	log.Printf("Checking authorization for user %s...", userID)
	authorized, err := a.checkUserAuthorization(userID, req.Namespace, req.Queue)

	if err != nil {
		// Internal error during authorization check
		log.Printf("ERROR: Failed to check authorization for user %s: %v", userID, err)
		http.Error(w, fmt.Sprintf("Failed to check authorization: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Authorization result: %t", authorized)
	log.Printf("User %s authorization check for namespace '%s', queue '%s': %t",
		userID, req.Namespace, req.Queue, authorized)

	// === STEP 4: PREPARE RESPONSE ===
	// Create a structured JSON response with all relevant information
	response := map[string]any{
		"authorized": authorized,                      // Boolean: is user authorized?
		"userID":     userID,                          // String: the authenticated user's ID
		"namespace":  req.Namespace,                   // String: the requested namespace
		"queue":      req.Queue,                       // String: the requested queue
		"timestamp":  time.Now().Format(time.RFC3339), // String: when this check was performed
	}

	log.Printf("Preparing response: %+v", response)

	// === STEP 5: SEND HTTP RESPONSE ===
	// Set the content type to indicate we're sending JSON
	w.Header().Set("Content-Type", "application/json")

	// Set the appropriate HTTP status code based on authorization result
	if authorized {
		log.Println("Setting response status: 200 OK")
		w.WriteHeader(http.StatusOK) // 200 = authorized
	} else {
		log.Println("Setting response status: 403 Forbidden")
		w.WriteHeader(http.StatusForbidden) // 403 = authenticated but not authorized
	}

	// Encode and send the JSON response
	log.Println("Encoding JSON response...")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode JSON response: %v", err)
		return
	}

	log.Printf("=== CHECK-AUTH HANDLER COMPLETE ===")
}
