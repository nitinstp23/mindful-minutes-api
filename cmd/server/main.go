package main

import (
	"log"

	"github.com/mindful-minutes/mindful-minutes-api/internal/http"
)

func main() {
	server, err := http.NewServer()
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	log.Println("Starting server...")
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
