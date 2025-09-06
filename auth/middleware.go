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

		// Check for required app role
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

		// Check for required app role
		hasRole := false
		if roles, ok := claims["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok && roleStr == "ServiceBus.DLQRetrigger" {
					hasRole = true
					log.Printf("User has required role: %s", roleStr)
					break
				}
			}
		}

		if !hasRole {
			log.Printf("User missing required role: ServiceBus.DLQRetrigger")
			log.Printf("Available roles: %v", claims["roles"])
			http.Error(w, "Insufficient permissions - missing ServiceBus.DLQRetrigger role", http.StatusForbidden)
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
