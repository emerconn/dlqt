package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("starting DLQT API")

	apiService := &APIService{}
	r := mux.NewRouter()

	r.HandleFunc("/trigger", apiService.handleTrigger).Methods("POST")
	r.HandleFunc("/fetch", apiService.handleFetch).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}
