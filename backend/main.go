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
	defer func() { _ = controlDB.DB.Close() }()

	// Initialize tenant connection manager
	tenantMgr := database.NewTenantManager(
		cfg.PostgresUser, cfg.PostgresPassword,
		cfg.PostgresHost, cfg.PostgresPort,
	)
	defer tenantMgr.Close()

	// Handlers
	authHandler := handlers.NewAuthHandler(cfg, controlDB)
	leagueHandler := handlers.NewLeagueHandler(cfg, controlDB)
	teamHandler := handlers.NewTeamHandler()
	swimmerHandler := handlers.NewSwimmerHandler()
	meetHandler := handlers.NewMeetHandler()
	eventHandler := handlers.NewEventHandler()
	timeHandler := handlers.NewTimeHandler()
	resultsHandler := handlers.NewResultsHandler(controlDB, tenantMgr)

	jwtSecret := []byte(cfg.JWTSecret)

	r := gin.Default()

	// Public routes
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/api/public/:leagueSlug/meets/:meetId", resultsHandler.GetPublicResults)

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
		// Teams
		leagues.GET("/teams", teamHandler.ListTeams)
		leagues.POST("/teams", teamHandler.CreateTeam)
		leagues.GET("/teams/:teamId", teamHandler.GetTeam)
		leagues.PUT("/teams/:teamId", teamHandler.UpdateTeam)
		leagues.DELETE("/teams/:teamId", teamHandler.DeleteTeam)

		// Swimmers
		leagues.GET("/swimmers", swimmerHandler.ListSwimmers)
		leagues.POST("/swimmers", swimmerHandler.CreateSwimmer)
		leagues.GET("/swimmers/:swimmerId", swimmerHandler.GetSwimmer)
		leagues.PUT("/swimmers/:swimmerId", swimmerHandler.UpdateSwimmer)
		leagues.DELETE("/swimmers/:swimmerId", swimmerHandler.DeleteSwimmer)

		// Meets
		leagues.GET("/meets", meetHandler.ListMeets)
		leagues.POST("/meets", meetHandler.CreateMeet)
		leagues.GET("/meets/:meetId", meetHandler.GetMeet)
		leagues.PUT("/meets/:meetId", meetHandler.UpdateMeet)
		leagues.DELETE("/meets/:meetId", meetHandler.DeleteMeet)

		// Events (nested under meets)
		leagues.GET("/meets/:meetId/events", eventHandler.ListEvents)
		leagues.POST("/meets/:meetId/events", eventHandler.CreateEvent)
		leagues.GET("/meets/:meetId/events/:eventId", eventHandler.GetEvent)
		leagues.PUT("/meets/:meetId/events/:eventId", eventHandler.UpdateEvent)
		leagues.DELETE("/meets/:meetId/events/:eventId", eventHandler.DeleteEvent)

		// Times (nested under events)
		leagues.GET("/meets/:meetId/events/:eventId/times", timeHandler.ListTimes)
		leagues.POST("/meets/:meetId/events/:eventId/times", timeHandler.CreateTime)
		leagues.GET("/meets/:meetId/events/:eventId/times/:timeId", timeHandler.GetTime)
		leagues.PUT("/meets/:meetId/events/:eventId/times/:timeId", timeHandler.UpdateTime)
		leagues.DELETE("/meets/:meetId/events/:eventId/times/:timeId", timeHandler.DeleteTime)
	}

	log.Printf("LapLogger backend starting on :%s", cfg.BackendPort)
	if err := r.Run(":" + cfg.BackendPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
