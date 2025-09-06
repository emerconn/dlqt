package main

type APIService struct{}

type AuthError struct {
	Code    int
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
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
