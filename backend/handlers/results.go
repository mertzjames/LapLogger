package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/database"
)

// ResultsHandler handles public (unauthenticated) results endpoints.
type ResultsHandler struct {
	controlDB *database.ControlDB
	tenantMgr *database.TenantManager
}

// NewResultsHandler creates a new ResultsHandler.
func NewResultsHandler(controlDB *database.ControlDB, tenantMgr *database.TenantManager) *ResultsHandler {
	return &ResultsHandler{controlDB: controlDB, tenantMgr: tenantMgr}
}

// publicEventResult is a single event with its times for public display.
type publicEventResult struct {
	ID         string              `json:"id"`
	Stroke     string              `json:"stroke"`
	Distance   int                 `json:"distance"`
	Unit       string              `json:"unit"`
	Gender     string              `json:"gender"`
	AgeGroup   string              `json:"age_group"`
	IsCustom   bool                `json:"is_custom"`
	CustomName string              `json:"custom_name,omitempty"`
	SortOrder  int                 `json:"sort_order"`
	Times      []publicTimeResult  `json:"times"`
}

// publicTimeResult is a single time entry for public display.
type publicTimeResult struct {
	SwimmerName    string `json:"swimmer_name"`
	TeamName       string `json:"team_name"`
	TimeHundredths int    `json:"time_hundredths"`
	IsExhibition   bool   `json:"is_exhibition"`
}

// publicMeetResult is the full public results response for a meet.
type publicMeetResult struct {
	MeetID     string              `json:"meet_id"`
	MeetName   string              `json:"meet_name"`
	Location   string              `json:"location"`
	MeetDate   interface{}         `json:"meet_date"`
	LeagueName string              `json:"league_name"`
	Events     []publicEventResult `json:"events"`
}

// GetPublicResults returns structured results for a public meet.
// GET /public/:leagueSlug/meets/:meetId
// No authentication required.
func (h *ResultsHandler) GetPublicResults(c *gin.Context) {
	leagueSlug := c.Param("leagueSlug")
	meetID := c.Param("meetId")

	if leagueSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing league slug"})
		return
	}
	if !isValidUUID(meetID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meet id"})
		return
	}

	// Look up the league by slug
	league, err := h.controlDB.GetLeagueBySlug(leagueSlug)
	if err != nil {
		log.Printf("Get league by slug error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve league"})
		return
	}
	if league == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "league not found"})
		return
	}

	// Connect to the tenant database
	tenantDB, err := h.tenantMgr.GetDB(league.DBName)
	if err != nil {
		log.Printf("Tenant DB connect error for %s: %v", league.DBName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to league database"})
		return
	}

	// Get the meet — must exist and be public
	var meetName, location string
	var meetDate interface{}
	var isPublic bool
	err = tenantDB.QueryRow(
		`SELECT name, location, meet_date, is_public FROM meets WHERE id = $1`, meetID,
	).Scan(&meetName, &location, &meetDate, &isPublic)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}
	if err != nil {
		log.Printf("Get meet error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get meet"})
		return
	}
	if !isPublic {
		c.JSON(http.StatusNotFound, gin.H{"error": "meet not found"})
		return
	}

	// Get all events for this meet
	eventRows, err := tenantDB.Query(`
		SELECT id, stroke, distance, unit, gender, age_group, is_custom, custom_name, sort_order
		FROM events WHERE meet_id = $1 ORDER BY sort_order, created_at
	`, meetID)
	if err != nil {
		log.Printf("List events for public results error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}
	defer func() { _ = eventRows.Close() }()

	var events []publicEventResult
	var eventIDs []string
	for eventRows.Next() {
		var e publicEventResult
		if err := eventRows.Scan(&e.ID, &e.Stroke, &e.Distance, &e.Unit, &e.Gender,
			&e.AgeGroup, &e.IsCustom, &e.CustomName, &e.SortOrder); err != nil {
			log.Printf("Scan event error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read events"})
			return
		}
		e.Times = []publicTimeResult{}
		events = append(events, e)
		eventIDs = append(eventIDs, e.ID)
	}
	if err := eventRows.Err(); err != nil {
		log.Printf("Event rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	// For each event, get times with swimmer + team names joined
	for i, evt := range events {
		timeRows, err := tenantDB.Query(`
			SELECT s.first_name || ' ' || s.last_name, COALESCE(t2.name, ''),
			       t.time_hundredths, t.is_exhibition
			FROM times t
			JOIN swimmers s ON s.id = t.swimmer_id
			LEFT JOIN teams t2 ON t2.id = s.team_id
			WHERE t.event_id = $1
			ORDER BY t.is_exhibition ASC, t.time_hundredths ASC
		`, evt.ID)
		if err != nil {
			log.Printf("List times for event %s error: %v", evt.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list times"})
			return
		}

		var times []publicTimeResult
		for timeRows.Next() {
			var tr publicTimeResult
			if err := timeRows.Scan(&tr.SwimmerName, &tr.TeamName,
				&tr.TimeHundredths, &tr.IsExhibition); err != nil {
				_ = timeRows.Close()
				log.Printf("Scan time error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read times"})
				return
			}
			times = append(times, tr)
		}
		if err := timeRows.Err(); err != nil {
			_ = timeRows.Close()
			log.Printf("Time rows error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list times"})
			return
		}
		_ = timeRows.Close()

		if times != nil {
			events[i].Times = times
		}
	}

	if events == nil {
		events = []publicEventResult{}
	}

	result := publicMeetResult{
		MeetID:     meetID,
		MeetName:   meetName,
		Location:   location,
		MeetDate:   meetDate,
		LeagueName: league.Name,
		Events:     events,
	}

	c.JSON(http.StatusOK, result)
}
