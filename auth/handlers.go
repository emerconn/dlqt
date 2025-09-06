package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func (a *AuthService) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	// If we reach here, user is already authenticated and authorized
	userID := r.Context().Value("userID").(string)

	var req CheckAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// User is authorized for anything since they have the ServiceBus.DLQRetrigger role
	response := map[string]any{
		"authorized": true,
		"userID":     userID,
		"namespace":  req.Namespace,
		"queue":      req.Queue,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
