// TODO: use better logs for both internal and HTTP

package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AuthMiddleware: %s %s", r.Method, r.URL)

		// extract token from Authorization header
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Missing or invalid Authorization header")
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			log.Printf("Empty token")
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		// JWKS URL using CMD tenant ID
		// TODO: remove hardcoded CMD tenant ID
		jwksURL := "https://login.microsoftonline.com/f09f69e2-b684-4c08-9195-f8f10f54154c/discovery/v2.0/keys"
		k, err := keyfunc.NewDefaultCtx(r.Context(), []string{jwksURL})
		if err != nil {
			log.Printf("Failed to create keyfunc: %v", err)
			http.Error(w, "Failed to fetch JWKS", http.StatusInternalServerError)
			return
		}

		// parse token with keyfunc (signature verification and standard claims)
		parsed, err := jwt.Parse(token, k.Keyfunc)
		if err != nil {
			log.Printf("token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// extract claims
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("failed to extract claims: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// validate audience claim
		// TODO: remove hardcoded API client ID
		if claims["aud"] != "074c5ac1-4ab2-4a8a-b811-2d7b8c4e419f" {
			log.Printf("invalid audience claim: %v", claims["aud"])
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// validate issuer claim
		// TODO: remove hardcoded CMD tenant ID
		if claims["iss"] != "https://login.microsoftonline.com/f09f69e2-b684-4c08-9195-f8f10f54154c/v2.0" {
			log.Printf("invalid issuer claim: %v", claims["iss"])
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// validate scope claim based on endpoint
		scp, ok := claims["scp"].(string)
		if !ok {
			log.Printf("invalid scope claim: %v", claims["scp"])
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		var requiredScope string
		switch r.URL.Path {
		case "/fetch":
			requiredScope = "dlq.read"
		case "/retrigger":
			requiredScope = "dlq.retrigger"
		default:
			log.Printf("unauthorized path: %s", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !strings.Contains(scp, requiredScope) {
			log.Printf("missing required scope %s in claim: %v", requiredScope, claims["scp"])
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Println("token validated successfully")

		// proceed to handler
		next.ServeHTTP(w, r)
	})
}
