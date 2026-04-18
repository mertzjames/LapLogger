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

// SwimmerHandler handles swimmer CRUD operations within a tenant database.
type SwimmerHandler struct{}

// NewSwimmerHandler creates a new SwimmerHandler.
func NewSwimmerHandler() *SwimmerHandler {
	return &SwimmerHandler{}
}

type createSwimmerRequest struct {
	TeamID      string `json:"team_id" binding:"required"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	DateOfBirth string `json:"date_of_birth" binding:"required"`
	Gender      string `json:"gender" binding:"required"`
}

type updateSwimmerRequest struct {
	TeamID      *string `json:"team_id"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender"`
}

// ListSwimmers returns all swimmers, optionally filtered by team_id query param.
func (h *SwimmerHandler) ListSwimmers(c *gin.Context) {
	db := middleware.GetTenantDB(c)
	teamID := c.Query("team_id")
	if teamID != "" && !isValidUUID(teamID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team_id"})
		return
	}

	var rows *sql.Rows
	var err error
	if teamID != "" {
		rows, err = db.Query(`
			SELECT id, team_id, first_name, last_name, date_of_birth, gender, created_at
			FROM swimmers WHERE team_id = $1 ORDER BY last_name, first_name
		`, teamID)
	} else {
		rows, err = db.Query(`
			SELECT id, team_id, first_name, last_name, date_of_birth, gender, created_at
			FROM swimmers ORDER BY last_name, first_name
		`)
	}
	if err != nil {
		log.Printf("List swimmers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list swimmers"})
		return
	}
	defer func() { _ = rows.Close() }()

	swimmers := []models.Swimmer{}
	for rows.Next() {
		var s models.Swimmer
		if err := rows.Scan(&s.ID, &s.TeamID, &s.FirstName, &s.LastName, &s.DateOfBirth, &s.Gender, &s.CreatedAt); err != nil {
			log.Printf("Scan swimmer error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read swimmer"})
			return
		}
		swimmers = append(swimmers, s)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list swimmers"})
		return
	}

	c.JSON(http.StatusOK, swimmers)
}

// GetSwimmer returns a single swimmer by ID.
func (h *SwimmerHandler) GetSwimmer(c *gin.Context) {
	swimmerID := c.Param("swimmerId")
	if !isValidUUID(swimmerID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swimmer id"})
		return
	}

	db := middleware.GetTenantDB(c)

	var s models.Swimmer
	err := db.QueryRow(`
		SELECT id, team_id, first_name, last_name, date_of_birth, gender, created_at
		FROM swimmers WHERE id = $1
	`, swimmerID).Scan(&s.ID, &s.TeamID, &s.FirstName, &s.LastName, &s.DateOfBirth, &s.Gender, &s.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "swimmer not found"})
		return
	}
	if err != nil {
		log.Printf("Get swimmer error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get swimmer"})
		return
	}

	c.JSON(http.StatusOK, s)
}

// CreateSwimmer creates a new swimmer.
func (h *SwimmerHandler) CreateSwimmer(c *gin.Context) {
	var req createSwimmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id, first_name, last_name, date_of_birth, and gender are required"})
		return
	}

	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	if firstName == "" || lastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first_name and last_name cannot be empty"})
		return
	}
	if len(firstName) > maxNameLength || len(lastName) > maxNameLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}
	if !isValidUUID(req.TeamID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team_id"})
		return
	}
	if !isValidDate(req.DateOfBirth) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date_of_birth must be YYYY-MM-DD"})
		return
	}

	gender := strings.ToUpper(strings.TrimSpace(req.Gender))
	if gender != "M" && gender != "F" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gender must be M or F"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Validate team exists
	var teamExists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`, req.TeamID).Scan(&teamExists)
	if err != nil {
		log.Printf("Check team error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate team"})
		return
	}
	if !teamExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team not found"})
		return
	}

	id := uuid.New().String()
	var s models.Swimmer
	err = db.QueryRow(`
		INSERT INTO swimmers (id, team_id, first_name, last_name, date_of_birth, gender)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, team_id, first_name, last_name, date_of_birth, gender, created_at
	`, id, req.TeamID, firstName, lastName, req.DateOfBirth, gender).Scan(
		&s.ID, &s.TeamID, &s.FirstName, &s.LastName, &s.DateOfBirth, &s.Gender, &s.CreatedAt,
	)
	if err != nil {
		log.Printf("Create swimmer error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create swimmer"})
		return
	}

	c.JSON(http.StatusCreated, s)
}

// UpdateSwimmer updates an existing swimmer.
func (h *SwimmerHandler) UpdateSwimmer(c *gin.Context) {
	var req updateSwimmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	swimmerID := c.Param("swimmerId")
	if !isValidUUID(swimmerID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swimmer id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Check swimmer exists
	var existing models.Swimmer
	err := db.QueryRow(`
		SELECT id, team_id, first_name, last_name, date_of_birth, gender, created_at
		FROM swimmers WHERE id = $1
	`, swimmerID).Scan(&existing.ID, &existing.TeamID, &existing.FirstName, &existing.LastName,
		&existing.DateOfBirth, &existing.Gender, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "swimmer not found"})
		return
	}
	if err != nil {
		log.Printf("Get swimmer for update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get swimmer"})
		return
	}

	// Apply updates
	if req.TeamID != nil {
		if !isValidUUID(*req.TeamID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team_id"})
			return
		}
		var teamExists bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`, *req.TeamID).Scan(&teamExists); err != nil {
			log.Printf("Check team error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate team"})
			return
		}
		if !teamExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "team not found"})
			return
		}
		existing.TeamID = *req.TeamID
	}
	if req.FirstName != nil {
		name := strings.TrimSpace(*req.FirstName)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "first_name cannot be empty"})
			return
		}
		existing.FirstName = name
	}
	if req.LastName != nil {
		name := strings.TrimSpace(*req.LastName)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "last_name cannot be empty"})
			return
		}
		existing.LastName = name
	}
	dobStr := ""
	if existing.DateOfBirth != nil {
		dobStr = existing.DateOfBirth.Format("2006-01-02")
	}
	if req.DateOfBirth != nil {
		if !isValidDate(*req.DateOfBirth) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date_of_birth must be YYYY-MM-DD"})
			return
		}
		dobStr = *req.DateOfBirth
	}
	if req.Gender != nil {
		gender := strings.ToUpper(strings.TrimSpace(*req.Gender))
		if gender != "M" && gender != "F" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gender must be M or F"})
			return
		}
		existing.Gender = gender
	}

	var s models.Swimmer
	err = db.QueryRow(`
		UPDATE swimmers SET team_id = $1, first_name = $2, last_name = $3, date_of_birth = $4, gender = $5
		WHERE id = $6
		RETURNING id, team_id, first_name, last_name, date_of_birth, gender, created_at
	`, existing.TeamID, existing.FirstName, existing.LastName, dobStr, existing.Gender, swimmerID).Scan(
		&s.ID, &s.TeamID, &s.FirstName, &s.LastName, &s.DateOfBirth, &s.Gender, &s.CreatedAt,
	)
	if err != nil {
		log.Printf("Update swimmer error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update swimmer"})
		return
	}

	c.JSON(http.StatusOK, s)
}

// DeleteSwimmer deletes a swimmer by ID.
func (h *SwimmerHandler) DeleteSwimmer(c *gin.Context) {
	swimmerID := c.Param("swimmerId")
	if !isValidUUID(swimmerID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swimmer id"})
		return
	}

	db := middleware.GetTenantDB(c)

	result, err := db.Exec(`DELETE FROM swimmers WHERE id = $1`, swimmerID)
	if err != nil {
		log.Printf("Delete swimmer error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete swimmer"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Rows affected error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete swimmer"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "swimmer not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
