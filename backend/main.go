package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/config"
	"github.com/laplogger/laplogger/database"
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

	_ = tenantMgr // will be used by tenant middleware in Phase 2

	r := gin.Default()

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("LapLogger backend starting on :%s", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
