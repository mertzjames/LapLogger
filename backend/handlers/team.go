package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laplogger/laplogger/middleware"
	"github.com/laplogger/laplogger/models"
)

// TeamHandler handles team CRUD operations within a tenant database.
type TeamHandler struct{}

// NewTeamHandler creates a new TeamHandler.
func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

type createTeamRequest struct {
	Name      string `json:"name" binding:"required"`
	ShortName string `json:"short_name"`
}

type updateTeamRequest struct {
	Name      *string `json:"name"`
	ShortName *string `json:"short_name"`
}

// ListTeams returns all teams in the tenant database.
func (h *TeamHandler) ListTeams(c *gin.Context) {
	db := middleware.GetTenantDB(c)

	rows, err := db.Query(`SELECT id, name, short_name, created_at FROM teams ORDER BY name`)
	if err != nil {
		log.Printf("List teams error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}
	defer func() { _ = rows.Close() }()

	teams := []models.Team{}
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.ShortName, &t.CreatedAt); err != nil {
			log.Printf("Scan team error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read team"})
			return
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GetTeam returns a single team by ID.
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID := c.Param("teamId")
	if !isValidUUID(teamID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	db := middleware.GetTenantDB(c)

	var t models.Team
	err := db.QueryRow(
		`SELECT id, name, short_name, created_at FROM teams WHERE id = $1`, teamID,
	).Scan(&t.ID, &t.Name, &t.ShortName, &t.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}
	if err != nil {
		log.Printf("Get team error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team"})
		return
	}

	c.JSON(http.StatusOK, t)
}

// CreateTeam creates a new team.
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req createTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
		return
	}
	if len(name) > maxNameLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}
	shortName := strings.TrimSpace(req.ShortName)

	db := middleware.GetTenantDB(c)
	id := uuid.New().String()
	var t models.Team
	err := db.QueryRow(`
		INSERT INTO teams (id, name, short_name) VALUES ($1, $2, $3)
		RETURNING id, name, short_name, created_at
	`, id, name, shortName).Scan(&t.ID, &t.Name, &t.ShortName, &t.CreatedAt)
	if err != nil {
		log.Printf("Create team error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// UpdateTeam updates an existing team.
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	var req updateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	db := middleware.GetTenantDB(c)
	teamID := c.Param("teamId")
	if !isValidUUID(teamID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	// Check team exists
	var existing models.Team
	err := db.QueryRow(
		`SELECT id, name, short_name, created_at FROM teams WHERE id = $1`, teamID,
	).Scan(&existing.ID, &existing.Name, &existing.ShortName, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}
	if err != nil {
		log.Printf("Get team for update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team"})
		return
	}

	// Apply updates
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
			return
		}
		if len(name) > maxNameLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
			return
		}
		existing.Name = name
	}
	if req.ShortName != nil {
		existing.ShortName = strings.TrimSpace(*req.ShortName)
	}

	var t models.Team
	err = db.QueryRow(`
		UPDATE teams SET name = $1, short_name = $2 WHERE id = $3
		RETURNING id, name, short_name, created_at
	`, existing.Name, existing.ShortName, teamID).Scan(&t.ID, &t.Name, &t.ShortName, &t.CreatedAt)
	if err != nil {
		log.Printf("Update team error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update team"})
		return
	}

	c.JSON(http.StatusOK, t)
}

// DeleteTeam deletes a team by ID.
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamID := c.Param("teamId")
	if !isValidUUID(teamID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	db := middleware.GetTenantDB(c)

	result, err := db.Exec(`DELETE FROM teams WHERE id = $1`, teamID)
	if err != nil {
		log.Printf("Delete team error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete team"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Rows affected error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete team"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
