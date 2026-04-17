package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/config"
	"github.com/laplogger/laplogger/database"
	"github.com/laplogger/laplogger/handlers"
	"github.com/laplogger/laplogger/middleware"
)

func main() {
	cfg := config.Load()

	// Connect to control plane DB and run migrations
	controlDB, err := database.NewControlDB(
		cfg.PostgresUser, cfg.PostgresPassword,
		cfg.PostgresHost, cfg.PostgresPort,
		cfg.PostgresDB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize control DB: %v", err)
	}
	defer controlDB.DB.Close()

	// Initialize tenant connection manager
	tenantMgr := database.NewTenantManager(
		cfg.PostgresUser, cfg.PostgresPassword,
		cfg.PostgresHost, cfg.PostgresPort,
	)
	defer tenantMgr.Close()

	// Handlers
	authHandler := handlers.NewAuthHandler(cfg, controlDB)
	leagueHandler := handlers.NewLeagueHandler(cfg, controlDB)

	jwtSecret := []byte(cfg.JWTSecret)

	r := gin.Default()

	// Public routes
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (no JWT required)
	auth := r.Group("/api/auth")
	{
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	// Authenticated routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired(jwtSecret))
	{
		api.GET("/me", authHandler.GetMe)
		api.POST("/leagues", leagueHandler.CreateLeague)
		api.GET("/leagues", leagueHandler.ListLeagues)
	}

	// Tenant-scoped routes (auth + tenant middleware)
	leagues := r.Group("/api/leagues/:leagueId")
	leagues.Use(middleware.AuthRequired(jwtSecret))
	leagues.Use(middleware.TenantRequired(controlDB, tenantMgr))
	{
		// Tenant CRUD endpoints will be added in Phase 4
	}

	log.Printf("LapLogger backend starting on :%s", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
