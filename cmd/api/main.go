package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting ActiveLog API...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to ActiveLog API"}`))

		log.Println("Server starting on :8080")
	})
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
