package main

import (
	"context"
	"log"

	backend "github.com/n1ckerr0r/chronograph-spb/backend"
	"github.com/n1ckerr0r/chronograph-spb/backend/internal/config"
	"github.com/n1ckerr0r/chronograph-spb/backend/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(context.Background(), db, backend.Migrations); err != nil {
		log.Fatal(err)
	}
	log.Print("database migrations are up to date")
}
