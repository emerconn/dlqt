package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

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
			log.Printf("Token key ID: %v", token.Header["kid"])    // Microsoft's key identifier

			// Validate that the token is using RSA256 algorithm
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Fetch the appropriate public key from Microsoft
			log.Println("Fetching Microsoft's public key for token validation...")
			return getPublicKeyForToken(token)
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
