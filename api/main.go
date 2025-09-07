package main

import (
	"log"
	"net/http"
)

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "fetch endpoint"}`))
}

func main() {
	log.Println("starting DLQT API")

	http.HandleFunc("/fetch", fetchHandler)

	log.Println("server starting on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
