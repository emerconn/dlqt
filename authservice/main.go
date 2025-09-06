package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"dlqt/internal/servicebus"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type AuthService struct {
	sbClient    *azservicebus.Client
	appObjectID string
}

type RetriggerRequest struct {
	Queue     string `json:"queue"`
	MessageID string `json:"messageId"`
}

func main() {
	// Load configuration from environment
	namespace := os.Getenv("AZURE_SERVICEBUS_NAMESPACE")
	appObjectID := os.Getenv("AZURE_APP_OBJECT_ID")

	if namespace == "" {
		log.Fatal("AZURE_SERVICEBUS_NAMESPACE must be set")
	}

	// Create Azure credentials for Service Bus (managed identity)
	sbCred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to create Service Bus credential: %v", err)
	}

	sbClient, err := azservicebus.NewClient(namespace, sbCred, nil)
	if err != nil {
		log.Fatalf("Failed to create Service Bus client: %v", err)
	}

	authService := &AuthService{
		sbClient:    sbClient,
		appObjectID: appObjectID,
	}

	// Setup HTTP server
	r := mux.NewRouter()
	r.Use(authMiddleware)

	r.HandleFunc("/retrigger", authService.handleRetrigger).Methods("POST")

	log.Println("Starting auth service on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// For Azure AD tokens, we need to validate against the public keys
			// This is a simplified version - in production, implement proper validation
			return []byte("secret"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract user ID from token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if oid, ok := claims["oid"].(string); ok {
				ctx := context.WithValue(r.Context(), "userID", oid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

func (a *AuthService) handleRetrigger(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// TODO: Check if user is authorized (e.g., member of specific group or app role)
	// For now, accept any valid Azure AD user
	log.Printf("User %s authorized for retrigger", userID)

	// Parse request
	var req RetriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrigger the message
	err := servicebus.RetriggerDeadLetterMessage(r.Context(), a.sbClient, req.Queue, req.MessageID)
	if err != nil {
		log.Printf("Failed to retrigger message: %v", err)
		http.Error(w, "Failed to retrigger message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (a *AuthService) checkUserAuthorization(userID string) (bool, error) {
	// TODO: Implement proper authorization check
	// This could check if user is in a specific Azure AD group
	// or has a specific app role assignment
	return true, nil
}
