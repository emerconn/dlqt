package main

import (
	"fmt"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, `{"namespace": "%s", "queue": "%s", "messageID": %s}`, namespace, queue, message.MessageID))
}
