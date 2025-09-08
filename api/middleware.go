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
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		log.Printf("Token: %s", token)

		// hardcoded JWKS URL
		jwksURL := "https://login.microsoftonline.com/f09f69e2-b684-4c08-9195-f8f10f54154c/discovery/v2.0/keys"
		k, err := keyfunc.NewDefaultCtx(r.Context(), []string{jwksURL})
		if err != nil {
			log.Printf("Failed to create keyfunc: %v", err)
			http.Error(w, "Failed to fetch JWKS", http.StatusInternalServerError)
			return
		}

		// parse token with signature verification using keyfunc
		_, err = jwt.Parse(token, k.Keyfunc)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Println("Token validated successfully")

		// proceed to handler
		next.ServeHTTP(w, r)
	})
}
