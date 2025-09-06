package main

// AuthService is the main service struct that handles authentication operations
// It doesn't have any fields yet, but could be extended with database connections,
// cache clients, or configuration in the future
type AuthService struct {
}

// CheckAuthRequest represents the JSON payload that clients send to check authorization
// This struct defines what data we expect in the request body
type CheckAuthRequest struct {
	Namespace string `json:"namespace"` // The Service Bus namespace to check access for
	Queue     string `json:"queue"`     // The specific queue within that namespace
}

// JWK represents a JSON Web Key from Microsoft's key endpoint
type JWK struct {
	Kty string `json:"kty"` // Key type
	Use string `json:"use"` // Key usage
	Kid string `json:"kid"` // Key ID
	N   string `json:"n"`   // RSA modulus
	E   string `json:"e"`   // RSA exponent
}

// JWKSet represents the response from Microsoft's JWKS endpoint
type JWKSet struct {
	Keys []JWK `json:"keys"`
}
