package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/router"
	"probig/pkg/config"
	"probig/pkg/crypto"
	"probig/pkg/database"
	pkgjwt "probig/pkg/jwt"
)

func main() {
	cfg := config.Load()

	crypto.Init(cfg.EncryptKey)
	pkgjwt.Init(cfg.JWTSecret)

	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	if err := database.AutoMigrate(models.AllModels...); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	r := gin.Default()
	router.Setup(r, cfg)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
		os.Exit(1)
	}
}
