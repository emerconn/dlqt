package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type AuthService struct {
}

type CheckAuthRequest struct {
	Namespace string `json:"namespace"`
	Queue     string `json:"queue"`
}

func main() {
	authService := &AuthService{}

	// Setup HTTP server
	r := mux.NewRouter()
	r.Use(authMiddleware)

	r.HandleFunc("/check-auth", authService.handleCheckAuth).Methods("POST")

	log.Println("Starting auth service on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// For Azure AD tokens, we need to validate against the public keys
			// This is a simplified version - in production, implement proper validation
			return []byte("secret"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract user ID from token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if oid, ok := claims["oid"].(string); ok {
				ctx := context.WithValue(r.Context(), "userID", oid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

func (a *AuthService) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Parse request
	var req CheckAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if user is authorized for this namespace and queue
	authorized, err := a.checkUserAuthorization(userID, req.Namespace, req.Queue)
	if err != nil {
		log.Printf("Failed to check authorization for user %s: %v", userID, err)
		http.Error(w, "Failed to check authorization", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s authorization check for namespace '%s', queue '%s': %t",
		userID, req.Namespace, req.Queue, authorized)

	response := map[string]any{
		"authorized": authorized,
		"userID":     userID,
		"namespace":  req.Namespace,
		"queue":      req.Queue,
	}

	w.Header().Set("Content-Type", "application/json")
	if authorized {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
	json.NewEncoder(w).Encode(response)
}

func (a *AuthService) checkUserAuthorization(userID, namespace, queue string) (bool, error) {
	// TODO: Implement proper authorization check
	// This could check if user is in a specific Azure AD group
	// or has a specific app role assignment
	// For now, we'll just return true for any valid Azure AD user
	log.Printf("Checking authorization for user %s on namespace %s, queue %s", userID, namespace, queue)
	return true, nil
}
