package main

import (
	"fmt"
	"log"
	"io/fs"
	"net/http"

	"github.com/medic666/probig/internal/auth"
	"github.com/medic666/probig/internal/config"
	"github.com/medic666/probig/internal/database"
	"github.com/medic666/probig/internal/handlers"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.EnsureUploadDir(cfg.Upload.Dir); err != nil {
		log.Fatalf("Failed to create upload dir: %v", err)
	}

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	if err := database.RunSeed(db); err != nil {
		log.Fatalf("Failed to run seed: %v", err)
	}
	log.Println("Database seed completed")

	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	var frontendFS fs.FS
	subFS, err := fs.Sub(frontendFiles, "web/dist")
	if err == nil {
		frontendFS = subFS
	}

	routerCfg := handlers.RouterConfig{
		DB:            db,
		JWTManager:    jwtMgr,
		UploadDir:     cfg.Upload.Dir,
		MaxUploadSize: cfg.Upload.MaxSize,
		FrontendFS:    frontendFS,
	}

	r := handlers.SetupRouter(routerCfg)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Default login: admin / admin123")

	log.Fatal(http.ListenAndServe(addr, r))
}
