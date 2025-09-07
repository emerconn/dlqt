package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Use string `json:"use"`
	Kty string `json:"kty"`
}

func getJWKS(tenantID string) (*JWKS, error) {
	url := fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", tenantID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}
	return &jwks, nil
}

func decodeRSAPublicKey(n, e string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}

	nInt := new(big.Int).SetBytes(nBytes)
	eInt := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: nInt,
		E: int(eInt.Int64()),
	}, nil
}

