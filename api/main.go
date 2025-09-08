package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("starting DLQT API")

	http.Handle("/fetch", AuthMiddleware(http.HandlerFunc(fetchHandler)))
	http.Handle("/retrig", AuthMiddleware(http.HandlerFunc(retriggerHandler)))

	log.Println("server starting on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
