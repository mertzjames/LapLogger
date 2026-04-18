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

// EventHandler handles event CRUD operations within a tenant database.
type EventHandler struct{}

// NewEventHandler creates a new EventHandler.
func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

// Valid strokes for standard events.
var validStrokes = map[string]bool{
	"Free": true, "Back": true, "Breast": true, "Fly": true, "IM": true,
}

// Valid units for events.
var validUnits = map[string]bool{
	"yards": true, "meters": true,
}

// Valid genders for events.
var validEventGenders = map[string]bool{
	"M": true, "F": true, "X": true,
}

type createEventRequest struct {
	Stroke     string `json:"stroke" binding:"required"`
	Distance   int    `json:"distance" binding:"required"`
	Unit       string `json:"unit"`
	Gender     string `json:"gender"`
	AgeGroup   string `json:"age_group"`
	IsCustom   bool   `json:"is_custom"`
	CustomName string `json:"custom_name"`
	SortOrder  int    `json:"sort_order"`
}

type updateEventRequest struct {
	Stroke     *string `json:"stroke"`
	Distance   *int    `json:"distance"`
	Unit       *string `json:"unit"`
	Gender     *string `json:"gender"`
	AgeGroup   *string `json:"age_group"`
	IsCustom   *bool   `json:"is_custom"`
	CustomName *string `json:"custom_name"`
	SortOrder  *int    `json:"sort_order"`
}

// ListEvents returns all events for a meet.
func (h *EventHandler) ListEvents(c *gin.Context) {
	meetID := c.Param("meetId")
	if !isValidUUID(meetID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meet id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify meet exists
	var meetExists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM meets WHERE id = $1)`, meetID).Scan(&meetExists); err != nil {
		log.Printf("Check meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate meet"})
		return
	}
	if !meetExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}

	rows, err := db.Query(`
		SELECT id, meet_id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order, created_at
		FROM events WHERE meet_id = $1 ORDER BY sort_order, created_at
	`, meetID)
	if err != nil {
		log.Printf("List events error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}
	defer func() { _ = rows.Close() }()

	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.MeetID, &e.Stroke, &e.Distance, &e.Unit, &e.Gender,
			&e.AgeGroup, &e.IsCustom, &e.CustomName, &e.SortOrder, &e.CreatedAt); err != nil {
			log.Printf("Scan event error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read event"})
			return
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetEvent returns a single event by ID.
func (h *EventHandler) GetEvent(c *gin.Context) {
	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	var e models.Event
	err := db.QueryRow(`
		SELECT id, meet_id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order, created_at
		FROM events WHERE id = $1 AND meet_id = $2
	`, eventID, meetID).Scan(&e.ID, &e.MeetID, &e.Stroke, &e.Distance, &e.Unit, &e.Gender,
		&e.AgeGroup, &e.IsCustom, &e.CustomName, &e.SortOrder, &e.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	if err != nil {
		log.Printf("Get event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get event"})
		return
	}

	c.JSON(http.StatusOK, e)
}

// CreateEvent creates a new event for a meet.
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stroke and distance are required"})
		return
	}

	stroke := strings.TrimSpace(req.Stroke)
	if !req.IsCustom && !validStrokes[stroke] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stroke must be one of: Free, Back, Breast, Fly, IM"})
		return
	}

	if req.Distance <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "distance must be positive"})
		return
	}

	unit := "yards"
	if req.Unit != "" {
		unit = strings.ToLower(strings.TrimSpace(req.Unit))
		if !validUnits[unit] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unit must be yards or meters"})
			return
		}
	}

	gender := "X"
	if req.Gender != "" {
		gender = strings.ToUpper(strings.TrimSpace(req.Gender))
		if !validEventGenders[gender] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gender must be M, F, or X"})
			return
		}
	}

	meetID := c.Param("meetId")
	if !isValidUUID(meetID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meet id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Verify meet exists
	var meetExists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM meets WHERE id = $1)`, meetID).Scan(&meetExists); err != nil {
		log.Printf("Check meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate meet"})
		return
	}
	if !meetExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}

	ageGroup := "Open"
	if req.AgeGroup != "" {
		ageGroup = strings.TrimSpace(req.AgeGroup)
	}

	customName := strings.TrimSpace(req.CustomName)

	id := uuid.New().String()
	var e models.Event
	err := db.QueryRow(`
		INSERT INTO events (id, meet_id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, meet_id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order, created_at
	`, id, meetID, stroke, req.Distance, unit, gender, ageGroup, req.IsCustom, customName, req.SortOrder).Scan(
		&e.ID, &e.MeetID, &e.Stroke, &e.Distance, &e.Unit, &e.Gender,
		&e.AgeGroup, &e.IsCustom, &e.CustomName, &e.SortOrder, &e.CreatedAt,
	)
	if err != nil {
		log.Printf("Create event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, e)
}

// UpdateEvent updates an existing event.
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	// Check event exists and belongs to this meet
	var existing models.Event
	err := db.QueryRow(`
		SELECT id, meet_id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order, created_at
		FROM events WHERE id = $1 AND meet_id = $2
	`, eventID, meetID).Scan(&existing.ID, &existing.MeetID, &existing.Stroke, &existing.Distance,
		&existing.Unit, &existing.Gender, &existing.AgeGroup, &existing.IsCustom,
		&existing.CustomName, &existing.SortOrder, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	if err != nil {
		log.Printf("Get event for update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get event"})
		return
	}

	// Apply updates
	if req.IsCustom != nil {
		existing.IsCustom = *req.IsCustom
	}
	if req.Stroke != nil {
		stroke := strings.TrimSpace(*req.Stroke)
		if !existing.IsCustom && !validStrokes[stroke] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stroke must be one of: Free, Back, Breast, Fly, IM"})
			return
		}
		existing.Stroke = stroke
	}
	if req.Distance != nil {
		if *req.Distance <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "distance must be positive"})
			return
		}
		existing.Distance = *req.Distance
	}
	if req.Unit != nil {
		unit := strings.ToLower(strings.TrimSpace(*req.Unit))
		if !validUnits[unit] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unit must be yards or meters"})
			return
		}
		existing.Unit = unit
	}
	if req.Gender != nil {
		gender := strings.ToUpper(strings.TrimSpace(*req.Gender))
		if !validEventGenders[gender] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gender must be M, F, or X"})
			return
		}
		existing.Gender = gender
	}
	if req.AgeGroup != nil {
		existing.AgeGroup = strings.TrimSpace(*req.AgeGroup)
	}
	if req.CustomName != nil {
		existing.CustomName = strings.TrimSpace(*req.CustomName)
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}

	var e models.Event
	err = db.QueryRow(`
		UPDATE events SET stroke = $1, distance = $2, unit = $3, gender = $4, age_group = $5,
		is_custom = $6, custom_name = $7, sort_order = $8
		WHERE id = $9 AND meet_id = $10
		RETURNING id, meet_id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order, created_at
	`, existing.Stroke, existing.Distance, existing.Unit, existing.Gender, existing.AgeGroup,
		existing.IsCustom, existing.CustomName, existing.SortOrder, eventID, meetID).Scan(
		&e.ID, &e.MeetID, &e.Stroke, &e.Distance, &e.Unit, &e.Gender,
		&e.AgeGroup, &e.IsCustom, &e.CustomName, &e.SortOrder, &e.CreatedAt,
	)
	if err != nil {
		log.Printf("Update event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
		return
	}

	c.JSON(http.StatusOK, e)
}

// DeleteEvent deletes an event by ID (cascades to times).
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	meetID := c.Param("meetId")
	eventID := c.Param("eventId")
	if !isValidUUID(meetID) || !isValidUUID(eventID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := middleware.GetTenantDB(c)

	result, err := db.Exec(`DELETE FROM events WHERE id = $1 AND meet_id = $2`, eventID, meetID)
	if err != nil {
		log.Printf("Delete event error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Rows affected error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
