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

func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("key type is not RSA: %s", jwk.Kty)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	if !e.IsInt64() {
		return nil, fmt.Errorf("exponent too large to fit in int")
	}

	eInt := e.Int64()
	if eInt > int64(^uint(0)>>1) {
		return nil, fmt.Errorf("exponent too large for int: %d", eInt)
	}

	return &rsa.PublicKey{N: n, E: int(eInt)}, nil
}

func getPublicKeyForToken(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("token does not have a key ID (kid)")
	}

	jwks, err := fetchMicrosoftJWKS()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Microsoft JWKS: %w", err)
	}

	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			return jwkToRSAPublicKey(jwk)
		}
	}

	return nil, fmt.Errorf("no key found with kid: %s", kid)
}
