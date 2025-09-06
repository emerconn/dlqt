package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5" // JWT library for parsing and validating JSON Web Tokens
	"github.com/gorilla/mux"       // HTTP router for handling different endpoints asdf
)

// AuthService is the main service struct that handles authentication operations
// It doesn't have any fields yet, but could be extended with database connections,
// cache clients, or configuration in the future
type AuthService struct {
}

// CheckAuthRequest represents the JSON payload that clients send to check authorization
// This struct defines what data we expect in the request body
type CheckAuthRequest struct {
	Namespace string `json:"namespace"` // The Service Bus namespace to check access for
	Queue     string `json:"queue"`     // The specific queue within that namespace
}

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

// authMiddleware is HTTP middleware that validates JWT tokens on every request
// Middleware in Go wraps HTTP handlers to add cross-cutting concerns like auth, logging, etc.
// It runs BEFORE the actual handler (like handleCheckAuth) and can block the request
func authMiddleware(next http.Handler) http.Handler {
	// Return a new handler function that wraps the original handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// === STEP 1: LOG REQUEST DETAILS ===
		// Log basic request information for debugging network issues
		log.Printf("=== AUTH MIDDLEWARE START ===")
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		log.Printf("Remote Address: %s", r.RemoteAddr)
		log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
		log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
		log.Printf("Content-Length: %s", r.Header.Get("Content-Length"))

		// Log all headers for debugging - helps identify missing or malformed headers
		log.Println("All request headers:")
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("  %s: %s", name, value)
			}
		}

		// === STEP 2: EXTRACT AUTHORIZATION HEADER ===
		// Get the Authorization header - this should contain "Bearer <JWT_TOKEN>"
		authHeader := r.Header.Get("Authorization")
		log.Printf("Authorization header: '%s'", authHeader)

		// Check if Authorization header exists
		if authHeader == "" {
			log.Println("ERROR: Authorization header missing")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return // Stop processing - request is rejected
		}

		// === STEP 3: VALIDATE BEARER TOKEN FORMAT ===
		// Extract the actual token by removing "Bearer " prefix
		log.Printf("Checking if header starts with 'Bearer '...")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// If TrimPrefix didn't change the string, it means "Bearer " wasn't there
		if tokenString == authHeader {
			log.Println("ERROR: Authorization header doesn't start with 'Bearer '")
			log.Printf("Raw header value: '%s'", authHeader)
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return // Stop processing - wrong format
		}

		// Log token info (safely - only show first 20 characters for security)
		log.Printf("Extracted token (first 20 chars): %s...", tokenString[:min(20, len(tokenString))])
		log.Printf("Full token length: %d characters", len(tokenString))

		// === STEP 4: PARSE AND VALIDATE JWT TOKEN ===
		log.Println("Parsing JWT token...")

		// Parse the JWT token using the jwt library
		// The callback function is called to provide the key for validation
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			log.Printf("JWT parsing callback called")
			log.Printf("Token algorithm: %v", token.Header["alg"]) // Shows signing algorithm (RS256, HS256, etc.)
			log.Printf("Token type: %v", token.Header["typ"])      // Should be "JWT"

			// ⚠️ SECURITY WARNING: This is using a hardcoded secret for validation
			// In production with Azure AD, you need to:
			// 1. Fetch Microsoft's public keys from https://login.microsoftonline.com/common/discovery/keys
			// 2. Validate the token was signed by Microsoft using those keys
			// 3. Validate audience, issuer, and expiration claims
			log.Println("WARNING: Using hardcoded secret for token validation - THIS IS NOT SECURE FOR PRODUCTION")
			return []byte("secret"), nil
		})

		// Check if token parsing failed
		if err != nil {
			log.Printf("ERROR: Failed to parse JWT token: %v", err)
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Check if token is valid (signature, expiration, etc.)
		if !token.Valid {
			log.Println("ERROR: JWT token is not valid")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Println("JWT token parsed successfully and is valid")

		// === STEP 5: EXTRACT USER INFORMATION FROM TOKEN CLAIMS ===
		// JWT tokens contain "claims" - key-value pairs with user/token information
		log.Println("Extracting claims from token...")
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			log.Printf("Token claims found: %d claims total", len(claims))

			// Log all claims for debugging - shows what data is available
			log.Println("All token claims:")
			for key, value := range claims {
				log.Printf("  %s: %v", key, value)
			}

			// Look for the "oid" claim - this is Azure AD's Object ID for the user
			// This uniquely identifies the user across all Azure AD applications
			if oid, ok := claims["oid"].(string); ok {
				log.Printf("Successfully extracted user ID (oid): %s", oid)

				// Add the user ID to the request context so handlers can access it
				// Context in Go is used to pass data between middleware and handlers
				ctx := context.WithValue(r.Context(), "userID", oid)
				log.Println("=== AUTH MIDDLEWARE SUCCESS - Proceeding to handler ===")

				// Continue to the actual handler with the enriched context
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				log.Println("ERROR: 'oid' claim not found or not a string")
				log.Printf("Available claims: %v", claims)
			}
		} else {
			log.Println("ERROR: Failed to extract claims from token")
		}

		// If we reach here, authentication failed
		log.Println("=== AUTH MIDDLEWARE FAILED ===")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

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

// checkUserAuthorization contains the business logic for determining if a user
// should be allowed to access a specific Service Bus namespace and queue
// This is where you would implement your actual authorization rules
func (a *AuthService) checkUserAuthorization(userID, namespace, queue string) (bool, error) {
	log.Printf("=== AUTHORIZATION CHECK START ===")
	log.Printf("Input parameters:")
	log.Printf("  User ID: '%s'", userID)      // Azure AD Object ID of the authenticated user
	log.Printf("  Namespace: '%s'", namespace) // Service Bus namespace they want to access
	log.Printf("  Queue: '%s'", queue)         // Specific queue within that namespace

	// 🚨 PLACEHOLDER IMPLEMENTATION 🚨
	// This is where you would implement your actual authorization logic, such as:
	//
	// 1. Check if user is in specific Azure AD groups:
	//    - Query Microsoft Graph API to get user's group memberships
	//    - Check if any of those groups have access to this namespace/queue
	//
	// 2. Check app role assignments:
	//    - Look at the JWT token's "roles" claim
	//    - Verify if user has roles like "ServiceBusReader", "ServiceBusWriter"
	//
	// 3. Database lookup:
	//    - Query your own database for user permissions
	//    - Match userID against namespace/queue access control lists
	//
	// 4. Policy-based access control:
	//    - Implement rules like "users from tenant X can access namespace Y"
	//    - Time-based access (business hours only)
	//    - IP-based restrictions
	//
	// 5. Integration with Azure RBAC:
	//    - Check if user has Azure Service Bus Data Owner/Receiver roles
	//    - Verify permissions through Azure Resource Manager API

	log.Println("WARNING: Using placeholder authorization logic - ALWAYS RETURNS TRUE")
	log.Println("TODO: Implement proper authorization check against Azure AD groups or app roles")
	log.Println("TODO: Consider integrating with:")
	log.Println("  - Microsoft Graph API for group membership")
	log.Println("  - Azure RBAC for Service Bus permissions")
	log.Println("  - Custom database for fine-grained access control")

	// For now, any authenticated user is authorized for any resource
	// This is obviously not secure for production use!
	result := true
	log.Printf("Authorization decision: %t", result)
	log.Printf("=== AUTHORIZATION CHECK COMPLETE ===")

	return result, nil
}
