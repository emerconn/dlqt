package main

type AuthService struct{}

type CheckAuthRequest struct {
	Namespace string `json:"namespace"`
	Queue     string `json:"queue"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSet struct {
	Keys []JWK `json:"keys"`
}
