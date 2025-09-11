package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"dlqt/internal/servicebus"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func respondSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: message})
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// extract query parameters
	namespace := r.URL.Query().Get("namespace")
	queue := r.URL.Query().Get("queue")
	slog.Info("received fetch request", "namespace", namespace, "queue", queue)

	// create service bus client
	client, err := servicebus.GetClient(namespace + ".servicebus.windows.net")
	if err != nil {
		slog.Error("failed to get service bus client", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get service bus client")
		return
	}

	// fetch dead letter message
	message, err := servicebus.FetchDeadLetterMessage(r.Context(), client, queue)
	if err != nil {
		slog.Error("failed to fetch dead letter message", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to fetch dead letter message")
		return
	}

	// get string of the first slice of bytes
	var messageBody string
	if len(message.Body) > 0 {
		messageBody = string(message.Body[0])
	}

	// map to JSON-serializable struct
	deadLetterMessage := &servicebus.DeadLetterMessage{
		Namespace:                  namespace,
		Queue:                      queue,
		MessageID:                  message.MessageID,
		Body:                       messageBody,
		ContentType:                message.ContentType,
		CorrelationID:              message.CorrelationID,
		DeadLetterErrorDescription: message.DeadLetterErrorDescription,
		DeadLetterReason:           message.DeadLetterReason,
		DeadLetterSource:           message.DeadLetterSource,
		DeliveryCount:              message.DeliveryCount,
		EnqueuedSequenceNumber:     message.EnqueuedSequenceNumber,
		EnqueuedTime:               message.EnqueuedTime,
		ExpiresAt:                  message.ExpiresAt,
		LockedUntil:                message.LockedUntil,
		PartitionKey:               message.PartitionKey,
		ReplyTo:                    message.ReplyTo,
		ReplyToSessionID:           message.ReplyToSessionID,
		ScheduledEnqueueTime:       message.ScheduledEnqueueTime,
		SequenceNumber:             message.SequenceNumber,
		SessionID:                  message.SessionID,
		State:                      int32(message.State),
		Subject:                    message.Subject,
		TimeToLive:                 message.TimeToLive,
		To:                         message.To,
		ApplicationProperties:      message.ApplicationProperties,
	}

	// convert to JSON
	jsonResponse, err := json.Marshal(deadLetterMessage)
	if err != nil {
		slog.Error("failed to marshal dead letter message to JSON", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to marshal dead letter message")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func retriggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// extract query parameters
	namespace := r.URL.Query().Get("namespace")
	queue := r.URL.Query().Get("queue")

	// extract messageID from body
	var requestBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		slog.Error("failed to decode JSON", "error", err)
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// extract messageID from body
	messageID, ok := requestBody["message-id"]
	if !ok {
		slog.Error("message-id not provided in request body")
		respondError(w, http.StatusBadRequest, "message-id not provided")
		return
	}

	slog.Info("received retrigger request", "namespace", namespace, "queue", queue, "messageID", messageID)

	client, err := servicebus.GetClient(namespace + ".servicebus.windows.net")
	if err != nil {
		slog.Error("failed to get service bus client", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get service bus client")
		return
	}

	err = servicebus.RetriggerDeadLetterMessage(r.Context(), client, queue, messageID)
	if err != nil {
		slog.Error("failed to retrigger dead letter message", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrigger message")
		return
	}

	// Send success response
	respondSuccess(w, fmt.Sprintf("message %s retriggered successfully", messageID))
}
