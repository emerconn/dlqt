package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AuthMiddleware: %s %s", r.Method, r.URL)

		// extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// basic validation
		if token == "" {
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		// TODO: implement full JWT validation here

		// parse token without validation to get header
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return nil, nil // Skip validation for now
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		kid, ok := parsedToken.Header["kid"].(string)
		if !ok {
			http.Error(w, "Missing key ID in token", http.StatusUnauthorized)
			return
		}

		// extract tenant from iss
		iss, ok := parsedToken.Claims.(jwt.MapClaims)["iss"].(string)
		if !ok {
			http.Error(w, "Missing iss in token", http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(iss, "https://login.microsoftonline.com/") {
			http.Error(w, "Invalid issuer", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(iss, "/")
		if len(parts) < 4 {
			http.Error(w, "Invalid issuer format", http.StatusUnauthorized)
			return
		}
		tenantID := parts[3]

		log.Printf("Token iss: %s, tenant: %s", iss, tenantID)
		log.Printf("Token kid: %s", kid)

		// fetch JWKS
		jwks, err := getJWKS(tenantID) // use dynamic tenant
		if err != nil {
			http.Error(w, "Failed to fetch JWKS", http.StatusInternalServerError)
			return
		}

		// find matching key
		var publicKey *rsa.PublicKey
		for _, key := range jwks.Keys {
			log.Printf("JWKS key kid: %s, use: %s", key.Kid, key.Use)
			if key.Kid == kid && key.Use == "sig" {
				publicKey, err = decodeRSAPublicKey(key.N, key.E)
				if err != nil {
					log.Printf("Failed to decode key %s: %v", key.Kid, err)
					continue
				}
				break
			}
		}
		if publicKey == nil {
			http.Error(w, "Failed to find matching key", http.StatusUnauthorized)
			return
		}

		// re-parse token with signature verification
		_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})
		if err != nil {
			http.Error(w, "Invalid token signature", http.StatusUnauthorized)
			return
		}

		log.Println("token signature verified successfully")

		// proceed to handler
		next.ServeHTTP(w, r)
	})
}
