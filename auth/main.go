package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type AuthService struct {
}

type CheckAuthRequest struct {
	Namespace string `json:"namespace"`
	Queue     string `json:"queue"`
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	log.Println("=== Starting DLQT Auth Service ===")
	log.Printf("Timestamp: %s", time.Now().Format(time.RFC3339))

	authService := &AuthService{}
	log.Println("Auth service instance created")

	// Setup HTTP server
	log.Println("Setting up HTTP router...")
	r := mux.NewRouter()

	log.Println("Adding auth middleware...")
	r.Use(authMiddleware)

	log.Println("Registering /check-auth endpoint...")
	r.HandleFunc("/check-auth", authService.handleCheckAuth).Methods("POST")

	log.Println("=== Auth service configuration complete ===")
	log.Println("Available endpoints:")
	log.Println("  POST /check-auth - Check user authorization for namespace/queue")
	log.Println("Starting auth service on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== AUTH MIDDLEWARE START ===")
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		log.Printf("Remote Address: %s", r.RemoteAddr)
		log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
		log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
		log.Printf("Content-Length: %s", r.Header.Get("Content-Length"))

		// Log all headers for debugging
		log.Println("All request headers:")
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("  %s: %s", name, value)
			}
		}

		authHeader := r.Header.Get("Authorization")
		log.Printf("Authorization header: '%s'", authHeader)

		if authHeader == "" {
			log.Println("ERROR: Authorization header missing")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		log.Printf("Checking if header starts with 'Bearer '...")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Println("ERROR: Authorization header doesn't start with 'Bearer '")
			log.Printf("Raw header value: '%s'", authHeader)
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		log.Printf("Extracted token (first 20 chars): %s...", tokenString[:min(20, len(tokenString))])
		log.Printf("Full token length: %d characters", len(tokenString))

		// Parse and validate JWT token
		log.Println("Parsing JWT token...")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			log.Printf("JWT parsing callback called")
			log.Printf("Token algorithm: %v", token.Header["alg"])
			log.Printf("Token type: %v", token.Header["typ"])
			// For Azure AD tokens, we need to validate against the public keys
			// This is a simplified version - in production, implement proper validation
			log.Println("WARNING: Using hardcoded secret for token validation - THIS IS NOT SECURE FOR PRODUCTION")
			return []byte("secret"), nil
		})

		if err != nil {
			log.Printf("ERROR: Failed to parse JWT token: %v", err)
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			log.Println("ERROR: JWT token is not valid")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Println("JWT token parsed successfully and is valid")

		// Extract user ID from token
		log.Println("Extracting claims from token...")
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			log.Printf("Token claims found: %d claims total", len(claims))

			// Log all claims for debugging
			log.Println("All token claims:")
			for key, value := range claims {
				log.Printf("  %s: %v", key, value)
			}

			if oid, ok := claims["oid"].(string); ok {
				log.Printf("Successfully extracted user ID (oid): %s", oid)
				ctx := context.WithValue(r.Context(), "userID", oid)
				log.Println("=== AUTH MIDDLEWARE SUCCESS - Proceeding to handler ===")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				log.Println("ERROR: 'oid' claim not found or not a string")
				log.Printf("Available claims: %v", claims)
			}
		} else {
			log.Println("ERROR: Failed to extract claims from token")
		}

		log.Println("=== AUTH MIDDLEWARE FAILED ===")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

func (a *AuthService) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== CHECK-AUTH HANDLER START ===")
	log.Printf("Request from: %s", r.RemoteAddr)

	userID := r.Context().Value("userID").(string)
	log.Printf("Authenticated user ID: %s", userID)

	// Parse request
	log.Println("Reading request body...")
	var req CheckAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Request decoded successfully:")
	log.Printf("  Namespace: '%s'", req.Namespace)
	log.Printf("  Queue: '%s'", req.Queue)

	// Check if user is authorized for this namespace and queue
	log.Printf("Checking authorization for user %s...", userID)
	authorized, err := a.checkUserAuthorization(userID, req.Namespace, req.Queue)
	if err != nil {
		log.Printf("ERROR: Failed to check authorization for user %s: %v", userID, err)
		http.Error(w, fmt.Sprintf("Failed to check authorization: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Authorization result: %t", authorized)
	log.Printf("User %s authorization check for namespace '%s', queue '%s': %t",
		userID, req.Namespace, req.Queue, authorized)

	response := map[string]any{
		"authorized": authorized,
		"userID":     userID,
		"namespace":  req.Namespace,
		"queue":      req.Queue,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	log.Printf("Preparing response: %+v", response)

	w.Header().Set("Content-Type", "application/json")
	if authorized {
		log.Println("Setting response status: 200 OK")
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("Setting response status: 403 Forbidden")
		w.WriteHeader(http.StatusForbidden)
	}

	log.Println("Encoding JSON response...")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode JSON response: %v", err)
		return
	}

	log.Printf("=== CHECK-AUTH HANDLER COMPLETE ===")
}

func (a *AuthService) checkUserAuthorization(userID, namespace, queue string) (bool, error) {
	log.Printf("=== AUTHORIZATION CHECK START ===")
	log.Printf("Input parameters:")
	log.Printf("  User ID: '%s'", userID)
	log.Printf("  Namespace: '%s'", namespace)
	log.Printf("  Queue: '%s'", queue)

	// TODO: Implement proper authorization check
	// This could check if user is in a specific Azure AD group
	// or has a specific app role assignment
	// For now, we'll just return true for any valid Azure AD user
	log.Println("WARNING: Using placeholder authorization logic - ALWAYS RETURNS TRUE")
	log.Println("TODO: Implement proper authorization check against Azure AD groups or app roles")

	result := true
	log.Printf("Authorization decision: %t", result)
	log.Printf("=== AUTHORIZATION CHECK COMPLETE ===")

	return result, nil
}
