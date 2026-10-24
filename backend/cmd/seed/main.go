package main

import (
	"log"
	"os"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
)

func main() {
	cfg := config.Load()
	sqlDB, err := db.Open(cfg.DBDriver, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	if err := seed.Run(sqlDB, cfg); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
