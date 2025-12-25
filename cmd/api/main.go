package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/valentinesamuel/activelog/internal/handlers"
)

func main() {
	fmt.Println("Starting ActiveLog API...")

	router := mux.NewRouter()

	healthHandler := handlers.NewHealthHandler()

	router.Handle("/health", healthHandler).Methods("GET")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to ActiveLog API"}`))

	}).Methods("GET")

	log.Println("Server starting on :8080")

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write([]byte(`{"message":"Welcome to ActiveLog API"}`))
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal(err)
	// }
	// })

	log.Fatal(http.ListenAndServe(":8080", router))

}
