package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"

	"dlqt/internal/servicebus"
)

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// extract query parameters
	namespace := r.URL.Query().Get("namespace")
	queue := r.URL.Query().Get("queue")
	log.Printf("received fetch request: namespace=%s, queue=%s", namespace, queue)

	// create service bus client
	client, err := servicebus.GetClient(namespace + ".servicebus.windows.net")
	if err != nil {
		log.Printf("failed to get service bus client: %v", err)
		http.Error(w, fmt.Sprintf("failed to get service bus client: %v", err), http.StatusInternalServerError)
		return
	}

	// fetch dead letter message
	message, err := servicebus.FetchDeadLetterMessage(r.Context(), client, queue)
	if err != nil {
		log.Printf("failed to fetch dead letter message: %v", err)
		http.Error(w, fmt.Sprintf("failed to fetch dead letter message: %v", err), http.StatusInternalServerError)
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
		log.Printf("failed to marshal dead letter message to JSON: %v", err)
		http.Error(w, fmt.Sprintf("failed to marshal dead letter message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func retriggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// extract query parameters
	namespace := r.URL.Query().Get("namespace")
	queue := r.URL.Query().Get("queue")

	// extract messageID from body
	var requestBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Printf("failed to decode JSON: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// extract messageID from body
	messageID, ok := requestBody["message-id"]
	if !ok {
		log.Printf("message-id not provided in request body")
		http.Error(w, "message-id not provided", http.StatusBadRequest)
		return
	}

	log.Printf("received retrigger request: namespace=%s, queue=%s, messageID=%s", namespace, queue, messageID)

	client, err := servicebus.GetClient(namespace + ".servicebus.windows.net")
	if err != nil {
		log.Printf("failed to get service bus client: %v", err)
		http.Error(w, fmt.Sprintf("failed to get service bus client: %v", err), http.StatusInternalServerError)
		return
	}

	err = servicebus.RetriggerDeadLetterMessage(r.Context(), client, queue, messageID)
	if err != nil {
		log.Printf("failed to retrigger dead letter message: %v", err)
		http.Error(w, fmt.Sprintf("failed to retrigger message: %v", err), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, "message %s retriggered successfully", html.EscapeString(messageID)))
}
