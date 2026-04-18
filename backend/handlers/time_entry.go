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

// TimeHandler handles time entry CRUD operations within a tenant database.
type TimeHandler struct{}

// NewTimeHandler creates a new TimeHandler.
func NewTimeHandler() *TimeHandler {
	return &TimeHandler{}
}

type createTimeRequest struct {
	SwimmerID      string `json:"swimmer_id" binding:"required"`
	TimeHundredths int    `json:"time_hundredths" binding:"required"`
	IsExhibition   bool   `json:"is_exhibition"`
}

type updateTimeRequest struct {
	SwimmerID      *string `json:"swimmer_id"`
	TimeHundredths *int    `json:"time_hundredths"`
	IsExhibition   *bool   `json:"is_exhibition"`
}

// ListTimes returns all time entries for an event.
func (h *TimeHandler) ListTimes(c *gin.Context) {
	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify event exists and belongs to this meet
	var eventExists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND meet_id = $2)`, eventID, meetID,
	).Scan(&eventExists); err != nil {
		log.Printf("Check event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate event"})
		return
	}
	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	rows, err := db.Query(`
		SELECT id, event_id, swimmer_id, time_hundredths, is_exhibition, created_at
		FROM times WHERE event_id = $1 ORDER BY time_hundredths
	`, eventID)
	if err != nil {
		log.Printf("List times error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list times"})
		return
	}
	defer func() { _ = rows.Close() }()

	times := []models.TimeEntry{}
	for rows.Next() {
		var t models.TimeEntry
		if err := rows.Scan(&t.ID, &t.EventID, &t.SwimmerID, &t.TimeHundredths, &t.IsExhibition, &t.CreatedAt); err != nil {
			log.Printf("Scan time error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read time"})
			return
		}
		times = append(times, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list times"})
		return
	}

	c.JSON(http.StatusOK, times)
}

// GetTime returns a single time entry by ID.
func (h *TimeHandler) GetTime(c *gin.Context) {
	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	timeID := c.Param("timeId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) || !isValidUUID(timeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify event belongs to meet
	var eventExists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND meet_id = $2)`, eventID, meetID,
	).Scan(&eventExists); err != nil {
		log.Printf("Check event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate event"})
		return
	}
	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	var t models.TimeEntry
	err := db.QueryRow(`
		SELECT id, event_id, swimmer_id, time_hundredths, is_exhibition, created_at
		FROM times WHERE id = $1 AND event_id = $2
	`, timeID, eventID).Scan(&t.ID, &t.EventID, &t.SwimmerID, &t.TimeHundredths, &t.IsExhibition, &t.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "time not found"})
		return
	}
	if err != nil {
		log.Printf("Get time error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get time"})
		return
	}

	c.JSON(http.StatusOK, t)
}

// CreateTime creates a new time entry for an event.
func (h *TimeHandler) CreateTime(c *gin.Context) {
	var req createTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "swimmer_id and time_hundredths are required"})
		return
	}

	if req.TimeHundredths <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time_hundredths must be positive"})
		return
	}
	if !isValidUUID(req.SwimmerID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swimmer_id"})
		return
	}

	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify event exists and belongs to this meet
	var eventExists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND meet_id = $2)`, eventID, meetID,
	).Scan(&eventExists); err != nil {
		log.Printf("Check event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate event"})
		return
	}
	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Verify swimmer exists
	var swimmerExists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM swimmers WHERE id = $1)`, req.SwimmerID).Scan(&swimmerExists); err != nil {
		log.Printf("Check swimmer error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate swimmer"})
		return
	}
	if !swimmerExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "swimmer not found"})
		return
	}

	id := uuid.New().String()
	var t models.TimeEntry
	err := db.QueryRow(`
		INSERT INTO times (id, event_id, swimmer_id, time_hundredths, is_exhibition)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, swimmer_id, time_hundredths, is_exhibition, created_at
	`, id, eventID, req.SwimmerID, req.TimeHundredths, req.IsExhibition).Scan(
		&t.ID, &t.EventID, &t.SwimmerID, &t.TimeHundredths, &t.IsExhibition, &t.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "time already exists for this swimmer in this event"})
			return
		}
		log.Printf("Create time error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create time"})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// UpdateTime updates an existing time entry.
func (h *TimeHandler) UpdateTime(c *gin.Context) {
	var req updateTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	timeID := c.Param("timeId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) || !isValidUUID(timeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify event belongs to meet
	var eventExists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND meet_id = $2)`, eventID, meetID,
	).Scan(&eventExists); err != nil {
		log.Printf("Check event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate event"})
		return
	}
	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Check time exists
	var existing models.TimeEntry
	err := db.QueryRow(`
		SELECT id, event_id, swimmer_id, time_hundredths, is_exhibition, created_at
		FROM times WHERE id = $1 AND event_id = $2
	`, timeID, eventID).Scan(&existing.ID, &existing.EventID, &existing.SwimmerID,
		&existing.TimeHundredths, &existing.IsExhibition, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "time not found"})
		return
	}
	if err != nil {
		log.Printf("Get time for update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get time"})
		return
	}

	// Apply updates
	if req.SwimmerID != nil {
		if !isValidUUID(*req.SwimmerID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swimmer_id"})
			return
		}
		var swimmerExists bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM swimmers WHERE id = $1)`, *req.SwimmerID).Scan(&swimmerExists); err != nil {
			log.Printf("Check swimmer error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate swimmer"})
			return
		}
		if !swimmerExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "swimmer not found"})
			return
		}
		existing.SwimmerID = *req.SwimmerID
	}
	if req.TimeHundredths != nil {
		if *req.TimeHundredths <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "time_hundredths must be positive"})
			return
		}
		existing.TimeHundredths = *req.TimeHundredths
	}
	if req.IsExhibition != nil {
		existing.IsExhibition = *req.IsExhibition
	}

	var t models.TimeEntry
	err = db.QueryRow(`
		UPDATE times SET swimmer_id = $1, time_hundredths = $2, is_exhibition = $3
		WHERE id = $4 AND event_id = $5
		RETURNING id, event_id, swimmer_id, time_hundredths, is_exhibition, created_at
	`, existing.SwimmerID, existing.TimeHundredths, existing.IsExhibition, timeID, eventID).Scan(
		&t.ID, &t.EventID, &t.SwimmerID, &t.TimeHundredths, &t.IsExhibition, &t.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "time already exists for this swimmer in this event"})
			return
		}
		log.Printf("Update time error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update time"})
		return
	}

	c.JSON(http.StatusOK, t)
}

// DeleteTime deletes a time entry by ID.
func (h *TimeHandler) DeleteTime(c *gin.Context) {
	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	timeID := c.Param("timeId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) || !isValidUUID(timeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify event belongs to meet
	var eventExists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND meet_id = $2)`, eventID, meetID,
	).Scan(&eventExists); err != nil {
		log.Printf("Check event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate event"})
		return
	}
	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	result, err := db.Exec(`DELETE FROM times WHERE id = $1 AND event_id = $2`, timeID, eventID)
	if err != nil {
		log.Printf("Delete time error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete time"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Rows affected error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete time"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "time not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// isDuplicateKeyError checks if a PostgreSQL error is a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505"))
}
