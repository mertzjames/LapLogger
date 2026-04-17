package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/config"
	"github.com/laplogger/laplogger/database"
	"github.com/laplogger/laplogger/models"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// LeagueHandler handles league (tenant) management.
type LeagueHandler struct {
	controlDB *database.ControlDB
	cfg       *config.Config
}

// NewLeagueHandler creates a new LeagueHandler.
func NewLeagueHandler(cfg *config.Config, controlDB *database.ControlDB) *LeagueHandler {
	return &LeagueHandler{controlDB: controlDB, cfg: cfg}
}

type createLeagueRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateLeague creates a new league, provisions the tenant database,
// and adds the requesting user as the league admin.
func (h *LeagueHandler) CreateLeague(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req createLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
		return
	}

	// Generate slug from name
	slug := slugRe.ReplaceAllString(strings.ToLower(name), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must contain alphanumeric characters"})
		return
	}

	// Create league record
	league, err := h.controlDB.CreateLeague(name, slug, "")
	if err != nil {
		log.Printf("Create league error: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "league name or slug already taken"})
		return
	}

	// Set the db_name based on the league ID (use sanitized form)
	dbName := fmt.Sprintf("laplogger_league_%s", strings.ReplaceAll(league.ID, "-", "_"))

	// Update the league record with the db_name
	if err := h.controlDB.UpdateLeagueDBName(league.ID, dbName); err != nil {
		log.Printf("Update league db_name error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to configure league"})
		return
	}
	league.DBName = dbName

	// Provision the tenant database
	if err := database.ProvisionTenantDB(
		h.controlDB.DB,
		h.cfg.PostgresUser, h.cfg.PostgresPassword,
		h.cfg.PostgresHost, h.cfg.PostgresPort,
		dbName,
	); err != nil {
		log.Printf("Provision tenant DB error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision league database"})
		return
	}

	// Add the creator as league admin
	if err := h.controlDB.AddMembership(userID.(string), league.ID, "admin"); err != nil {
		log.Printf("Add membership error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add membership"})
		return
	}

	c.JSON(http.StatusCreated, league)
}

// ListLeagues returns all leagues the authenticated user belongs to.
func (h *LeagueHandler) ListLeagues(c *gin.Context) {
	userID, _ := c.Get("userID")

	leagues, err := h.controlDB.GetUserLeagues(userID.(string))
	if err != nil {
		log.Printf("List leagues error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list leagues"})
		return
	}

	if leagues == nil {
		leagues = []models.League{}
	}

	c.JSON(http.StatusOK, leagues)
}
