package main

import (
	"log"

	"github.com/mindful-minutes/mindful-minutes-api/internal/http"
)

func main() {
	server := http.NewServer()
	
	log.Println("Starting server...")
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}