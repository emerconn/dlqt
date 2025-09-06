package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Bearer token
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

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getPublicKeyForToken(token)
		})

		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check for required scopes (dlq.fetch or dlq.retrigger)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Check app audience (should be our app's client ID)
		expectedAudience := "795ab387-3200-4ae3-81e3-a096b787e155"
		if aud, ok := claims["aud"].(string); !ok || aud != expectedAudience {
			log.Printf("Wrong audience: got %v, expected %s", claims["aud"], expectedAudience)
			http.Error(w, "Token not for this application", http.StatusForbidden)
			return
		}

		// Check for required scopes (dlq.fetch or dlq.retrigger)
		hasRequiredScope := false
		if scp, ok := claims["scp"].(string); ok {
			scopes := strings.Fields(scp)
			for _, scope := range scopes {
				if scope == "dlq.fetch" || scope == "dlq.retrigger" {
					hasRequiredScope = true
					log.Printf("User has required scope: %s", scope)
					break
				}
			}
		}

		if !hasRequiredScope {
			log.Printf("User missing required scopes: dlq.fetch or dlq.retrigger")
			log.Printf("Available scopes: %v", claims["scp"])
			http.Error(w, "Insufficient permissions - missing required scope (dlq.fetch or dlq.retrigger)", http.StatusForbidden)
			return
		}

		// Get user ID for logging
		userID := "unknown"
		if oid, ok := claims["oid"].(string); ok {
			userID = oid
		}

		log.Printf("Authorized user: %s", userID)
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
