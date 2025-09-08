package main

import (
	"fmt"
	"log"
	"net/http"
)

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract query parameters
	namespace := r.URL.Query().Get("namespace")
	queue := r.URL.Query().Get("queue")

	// Log them
	log.Printf("Received fetch request: namespace=%s, queue=%s", namespace, queue)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, `{"message": "fetch endpoint", "namespace": "%s", "queue": "%s"}`, namespace, queue))
}
