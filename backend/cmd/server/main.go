package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	_ "github.com/yumeng-tech/ai-collab-board/backend/internal/modules/botmem"
	_ "github.com/yumeng-tech/ai-collab-board/backend/internal/modules/timebilling"
	_ "github.com/yumeng-tech/ai-collab-board/backend/internal/modules/wiki"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/server"
)

func main() {
	cfg := config.Load()
	sqlDB, err := db.Open(cfg.DBDriver, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer sqlDB.Close()

	if os.Getenv("AUTO_SEED") != "0" {
		if err := seed.Run(sqlDB, cfg); err != nil {
			log.Printf("seed warning: %v", err)
		}
	}

	handler := server.NewRouter(sqlDB, cfg)
	addr := ":" + cfg.Port
	log.Printf("AI 协作看板 API listening on %s (env=%s)", addr, cfg.AppEnv)
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
