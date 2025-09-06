package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// fetchMicrosoftJWKS fetches the public keys from Microsoft's JWKS endpoint
func fetchMicrosoftJWKS() (*JWKSet, error) {
	resp, err := http.Get("https://login.microsoftonline.com/common/discovery/v2.0/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	return &jwks, nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Validate that this is an RSA key
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("key type is not RSA: %s", jwk.Kty)
	}

	// Decode the modulus (n) from base64url
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent (e) from base64url
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Validate that exponent is not too large for int
	if !e.IsInt64() {
		return nil, fmt.Errorf("exponent too large to fit in int")
	}

	eInt := e.Int64()
	if eInt > int64(^uint(0)>>1) { // Check if it fits in int
		return nil, fmt.Errorf("exponent too large for int: %d", eInt)
	}

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(eInt),
	}

	return publicKey, nil
}

// getPublicKeyForToken fetches the appropriate public key for validating a JWT token
func getPublicKeyForToken(token *jwt.Token) (interface{}, error) {
	// Validate the signing method
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	// Get the key ID from the token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("token does not have a key ID (kid)")
	}

	// Fetch Microsoft's public keys
	jwks, err := fetchMicrosoftJWKS()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Microsoft JWKS: %w", err)
	}

	// Find the key with matching kid
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			publicKey, err := jwkToRSAPublicKey(jwk)
			if err != nil {
				return nil, fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
			}
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("no key found with kid: %s", kid)
}
