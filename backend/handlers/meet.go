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

// MeetHandler handles meet CRUD operations within a tenant database.
type MeetHandler struct{}

// NewMeetHandler creates a new MeetHandler.
func NewMeetHandler() *MeetHandler {
	return &MeetHandler{}
}

type createMeetRequest struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location"`
	MeetDate string `json:"meet_date" binding:"required"`
	IsPublic *bool  `json:"is_public"`
}

type updateMeetRequest struct {
	Name     *string `json:"name"`
	Location *string `json:"location"`
	MeetDate *string `json:"meet_date"`
	IsPublic *bool   `json:"is_public"`
}

// ListMeets returns all meets in the tenant database.
func (h *MeetHandler) ListMeets(c *gin.Context) {
	db := middleware.GetTenantDB(c)

	rows, err := db.Query(`SELECT id, name, location, meet_date, is_public, created_at FROM meets ORDER BY meet_date DESC`)
	if err != nil {
		log.Printf("List meets error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list meets"})
		return
	}
	defer func() { _ = rows.Close() }()

	meets := []models.Meet{}
	for rows.Next() {
		var m models.Meet
		if err := rows.Scan(&m.ID, &m.Name, &m.Location, &m.MeetDate, &m.IsPublic, &m.CreatedAt); err != nil {
			log.Printf("Scan meet error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read meet"})
			return
		}
		meets = append(meets, m)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list meets"})
		return
	}

	c.JSON(http.StatusOK, meets)
}

// GetMeet returns a single meet by ID.
func (h *MeetHandler) GetMeet(c *gin.Context) {
	meetID := c.Param("meetId")
	if !isValidUUID(meetID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meet id"})
		return
	}

	db := middleware.GetTenantDB(c)

	var m models.Meet
	err := db.QueryRow(
		`SELECT id, name, location, meet_date, is_public, created_at FROM meets WHERE id = $1`, meetID,
	).Scan(&m.ID, &m.Name, &m.Location, &m.MeetDate, &m.IsPublic, &m.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}
	if err != nil {
		log.Printf("Get meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get meet"})
		return
	}

	c.JSON(http.StatusOK, m)
}

// CreateMeet creates a new meet.
func (h *MeetHandler) CreateMeet(c *gin.Context) {
	var req createMeetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and meet_date are required"})
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
	location := strings.TrimSpace(req.Location)
	if len(location) > maxLocationLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location too long"})
		return
	}
	if !isValidDate(req.MeetDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "meet_date must be YYYY-MM-DD"})
		return
	}
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	db := middleware.GetTenantDB(c)
	id := uuid.New().String()
	var m models.Meet
	err := db.QueryRow(`
		INSERT INTO meets (id, name, location, meet_date, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, location, meet_date, is_public, created_at
	`, id, name, location, req.MeetDate, isPublic).Scan(
		&m.ID, &m.Name, &m.Location, &m.MeetDate, &m.IsPublic, &m.CreatedAt,
	)
	if err != nil {
		log.Printf("Create meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create meet"})
		return
	}

	c.JSON(http.StatusCreated, m)
}

// UpdateMeet updates an existing meet.
func (h *MeetHandler) UpdateMeet(c *gin.Context) {
	var req updateMeetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	meetID := c.Param("meetId")
	if !isValidUUID(meetID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meet id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Check meet exists
	var existing models.Meet
	err := db.QueryRow(
		`SELECT id, name, location, meet_date, is_public, created_at FROM meets WHERE id = $1`, meetID,
	).Scan(&existing.ID, &existing.Name, &existing.Location, &existing.MeetDate, &existing.IsPublic, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}
	if err != nil {
		log.Printf("Get meet for update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get meet"})
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
	if req.Location != nil {
		loc := strings.TrimSpace(*req.Location)
		if len(loc) > maxLocationLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": "location too long"})
			return
		}
		existing.Location = loc
	}
	meetDateStr := ""
	if existing.MeetDate != nil {
		meetDateStr = existing.MeetDate.Format("2006-01-02")
	}
	if req.MeetDate != nil {
		if !isValidDate(*req.MeetDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "meet_date must be YYYY-MM-DD"})
			return
		}
		meetDateStr = *req.MeetDate
	}
	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}

	var m models.Meet
	err = db.QueryRow(`
		UPDATE meets SET name = $1, location = $2, meet_date = $3, is_public = $4
		WHERE id = $5
		RETURNING id, name, location, meet_date, is_public, created_at
	`, existing.Name, existing.Location, meetDateStr, existing.IsPublic, meetID).Scan(
		&m.ID, &m.Name, &m.Location, &m.MeetDate, &m.IsPublic, &m.CreatedAt,
	)
	if err != nil {
		log.Printf("Update meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update meet"})
		return
	}

	c.JSON(http.StatusOK, m)
}

// DeleteMeet deletes a meet by ID (cascades to events and times).
func (h *MeetHandler) DeleteMeet(c *gin.Context) {
	meetID := c.Param("meetId")
	if !isValidUUID(meetID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meet id"})
		return
	}

	db := middleware.GetTenantDB(c)

	result, err := db.Exec(`DELETE FROM meets WHERE id = $1`, meetID)
	if err != nil {
		log.Printf("Delete meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete meet"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Rows affected error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete meet"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
