package main

import (
	"log"

	"documentservice/internal/config"
	"documentservice/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
