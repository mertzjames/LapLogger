package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// The control DB operations (league record + membership) are atomic.
// If tenant DB provisioning fails, the control DB records are rolled back.
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

	// Generate the db_name up front so it can be inserted atomically
	leagueUUID := uuid.New().String()
	dbName := fmt.Sprintf("laplogger_league_%s", strings.ReplaceAll(leagueUUID, "-", "_"))

	// Atomically create league record + membership in a single transaction
	league, err := h.controlDB.CreateLeagueWithMembership(name, slug, dbName, userID.(string), "admin")
	if err != nil {
		log.Printf("Create league error: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "league name or slug already taken"})
		return
	}

	// Provision the tenant database
	if err := database.ProvisionTenantDB(
		h.controlDB.DB,
		h.cfg.PostgresUser, h.cfg.PostgresPassword,
		h.cfg.PostgresHost, h.cfg.PostgresPort,
		dbName,
	); err != nil {
		log.Printf("Provision tenant DB error: %v", err)
		// Rollback: remove the league + membership from the control DB
		if delErr := h.controlDB.DeleteLeague(league.ID); delErr != nil {
			log.Printf("Rollback delete league error: %v", delErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision league database"})
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
