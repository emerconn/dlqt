package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dlqt/internal/servicebus"
)

// MessageResponse represents the JSON-serializable version of a Service Bus message
type MessageResponse struct {
	Namespace                  string                 `json:"namespace"`
	Queue                      string                 `json:"queue"`
	MessageID                  string                 `json:"messageID"`
	Body                       string                 `json:"body"` // Changed from []string to string
	ContentType                *string                `json:"contentType,omitempty"`
	CorrelationID              *string                `json:"correlationID,omitempty"`
	DeadLetterErrorDescription *string                `json:"deadLetterErrorDescription,omitempty"`
	DeadLetterReason           *string                `json:"deadLetterReason,omitempty"`
	DeadLetterSource           *string                `json:"deadLetterSource,omitempty"`
	DeliveryCount              uint32                 `json:"deliveryCount"`
	EnqueuedSequenceNumber     *int64                 `json:"enqueuedSequenceNumber,omitempty"`
	EnqueuedTime               *time.Time             `json:"enqueuedTime,omitempty"`
	ExpiresAt                  *time.Time             `json:"expiresAt,omitempty"`
	LockedUntil                *time.Time             `json:"lockedUntil,omitempty"`
	PartitionKey               *string                `json:"partitionKey,omitempty"`
	ReplyTo                    *string                `json:"replyTo,omitempty"`
	ReplyToSessionID           *string                `json:"replyToSessionID,omitempty"`
	ScheduledEnqueueTime       *time.Time             `json:"scheduledEnqueueTime,omitempty"`
	SequenceNumber             *int64                 `json:"sequenceNumber,omitempty"`
	SessionID                  *string                `json:"sessionID,omitempty"`
	State                      int32                  `json:"state"`
	Subject                    *string                `json:"subject,omitempty"`
	TimeToLive                 *time.Duration         `json:"timeToLive,omitempty"`
	To                         *string                `json:"to,omitempty"`
	ApplicationProperties      map[string]interface{} `json:"applicationProperties,omitempty"`
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// extract query parameters
	namespace := r.URL.Query().Get("namespace")
	queue := r.URL.Query().Get("queue")
	log.Printf("received fetch request: namespace=%s, queue=%s", namespace, queue)

	client, err := servicebus.GetClient(namespace + ".servicebus.windows.net")
	if err != nil {
		log.Printf("failed to get service bus client: %v", err)
		http.Error(w, fmt.Sprintf("failed to get service bus client: %v", err), http.StatusInternalServerError)
		return
	}

	message, err := servicebus.FetchDeadLetterMessage(r.Context(), client, queue)
	if err != nil {
		log.Printf("failed to fetch dead letter message: %v", err)
		http.Error(w, fmt.Sprintf("failed to fetch dead letter message: %v", err), http.StatusInternalServerError)
		return
	}

	// get string of the first slice of bytes
	var bodyString string
	if len(message.Body) > 0 {
		bodyString = string(message.Body[0])
	}

	response := MessageResponse{
		Namespace:                  namespace,
		Queue:                      queue,
		MessageID:                  message.MessageID,
		Body:                       bodyString,
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

	// Convert to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("failed to marshal message to JSON: %v", err)
		http.Error(w, fmt.Sprintf("failed to marshal message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
