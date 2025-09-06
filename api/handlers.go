package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// validateTokenWithScope validates JWT token and checks for required scope
func validateTokenWithScope(r *http.Request, requiredScope string) (string, error) {
	// Get Bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", &AuthError{Code: http.StatusUnauthorized, Message: "Authorization header required"}
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", &AuthError{Code: http.StatusUnauthorized, Message: "Bearer token required"}
	}

	// Parse and validate JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return getPublicKeyForToken(token)
	})

	if err != nil || !token.Valid {
		log.Printf("Invalid token: %v", err)
		return "", &AuthError{Code: http.StatusUnauthorized, Message: "Invalid token"}
	}

	// Check for required scopes
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", &AuthError{Code: http.StatusUnauthorized, Message: "Invalid token claims"}
	}

	// Check app audience (should be our app's client ID)
	expectedAudience := "795ab387-3200-4ae3-81e3-a096b787e155"
	if aud, ok := claims["aud"].(string); !ok || aud != expectedAudience {
		log.Printf("Wrong audience: got %v, expected %s", claims["aud"], expectedAudience)
		return "", &AuthError{Code: http.StatusForbidden, Message: "Token not for this application"}
	}

	// Check for required scope
	hasRequiredScope := false
	if scp, ok := claims["scp"].(string); ok {
		scopes := strings.Fields(scp)
		for _, scope := range scopes {
			if scope == requiredScope {
				hasRequiredScope = true
				log.Printf("User has required scope: %s", scope)
				break
			}
		}
	}

	if !hasRequiredScope {
		log.Printf("User missing required scope: %s", requiredScope)
		log.Printf("Available scopes: %v", claims["scp"])
		return "", &AuthError{Code: http.StatusForbidden, Message: "Insufficient permissions - missing required scope: " + requiredScope}
	}

	// Get user ID for logging
	userID := "unknown"
	if oid, ok := claims["oid"].(string); ok {
		userID = oid
	}

	log.Printf("Authorized user: %s with scope: %s", userID, requiredScope)
	return userID, nil
}

func (a *APIService) handleTrigger(w http.ResponseWriter, r *http.Request) {
	userID, err := validateTokenWithScope(r, "dlq.retrigger")
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			http.Error(w, authErr.Message, authErr.Code)
		} else {
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
		}
		return
	}

	// User is authorized with dlq.retrigger scope
	response := map[string]any{
		"authorized": true,
		"userID":     userID,
		"scope":      "dlq.retrigger",
		"action":     "trigger",
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (a *APIService) handleFetch(w http.ResponseWriter, r *http.Request) {
	userID, err := validateTokenWithScope(r, "dlq.fetch")
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			http.Error(w, authErr.Message, authErr.Code)
		} else {
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
		}
		return
	}

	// User is authorized with dlq.fetch scope
	response := map[string]any{
		"authorized": true,
		"userID":     userID,
		"scope":      "dlq.fetch",
		"action":     "fetch",
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
