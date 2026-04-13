package main

import (
	"log"

	"clinic-api/internal/config"
	"clinic-api/internal/database"
	"clinic-api/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	srv := server.New(cfg, db)
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(srv.Start())
}
