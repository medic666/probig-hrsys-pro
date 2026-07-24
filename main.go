package main

import (
	"embed"
	"log"
	"os"

	"probig/internal/config"
	"probig/internal/database"
	"probig/internal/router"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

func main() {
	cfg := config.Load()

	if cfg.DevMode {
		os.MkdirAll("frontend/dist", 0755)
	}

	database.Init(cfg.DBPath)
	database.Seed()

	r := router.Setup(cfg, frontendFS)

	log.Printf("Server starting on :%s (dev mode: %v)", cfg.Port, cfg.DevMode)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
