package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("starting DLQT API")

	http.HandleFunc("/fetch", fetchHandler)

	log.Println("server starting on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
