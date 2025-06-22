package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"laplogger/models"
)

type TimeHandler struct {
	db *sql.DB
}

func NewTimeHandler(db *sql.DB) *TimeHandler {
	return &TimeHandler{db: db}
}

func (h *TimeHandler) CreateTime(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SwimmerID <= 0 || req.EventID <= 0 || req.TimeMs <= 0 {
		http.Error(w, "Swimmer ID, Event ID, and Time are required", http.StatusBadRequest)
		return
	}

	// Insert the time
	result, err := h.db.Exec(`
		INSERT INTO swim_times (swimmer_id, event_id, meet_id, time_ms, notes) 
		VALUES (?, ?, ?, ?, ?)`,
		req.SwimmerID, req.EventID, req.MeetID, req.TimeMs, req.Notes)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created time with details
	timeWithDetails, err := h.getTimeWithDetails(int(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(timeWithDetails)
}

func (h *TimeHandler) GetTimesBySwimmer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swimmerID, err := strconv.Atoi(vars["swimmer_id"])
	if err != nil {
		http.Error(w, "Invalid swimmer ID", http.StatusBadRequest)
		return
	}

	times, err := h.getTimesWithDetailsBySwimmer(swimmerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(times)
}

func (h *TimeHandler) GetAllTimes(w http.ResponseWriter, r *http.Request) {
	times, err := h.getAllTimesWithDetails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(times)
}

// Helper function to get a single time with details
func (h *TimeHandler) getTimeWithDetails(timeID int) (*models.SwimTimeWithDetails, error) {
	query := `
		SELECT 
			st.id, st.swimmer_id, st.event_id, st.meet_id, st.time_ms, st.notes, st.recorded_at,
			s.name as swimmer_name,
			e.name as event_name,
			str.name as stroke_name,
			e.distance,
			m.name as meet_name
		FROM swim_times st
		JOIN swimmers s ON st.swimmer_id = s.id
		JOIN events e ON st.event_id = e.id
		JOIN strokes str ON e.stroke_id = str.id
		LEFT JOIN meets m ON st.meet_id = m.id
		WHERE st.id = ?
	`

	var timeDetails models.SwimTimeWithDetails
	var meetName sql.NullString

	err := h.db.QueryRow(query, timeID).Scan(
		&timeDetails.ID, &timeDetails.SwimmerID, &timeDetails.EventID, 
		&timeDetails.MeetID, &timeDetails.TimeMs, &timeDetails.Notes, &timeDetails.RecordedAt,
		&timeDetails.SwimmerName, &timeDetails.EventName, &timeDetails.StrokeName,
		&timeDetails.Distance, &meetName,
	)

	if err != nil {
		return nil, err
	}

	if meetName.Valid {
		timeDetails.MeetName = &meetName.String
	}

	// Format the time
	timeDetails.FormattedTime = timeDetails.SwimTime.FormatTime()

	return &timeDetails, nil
}

// Helper function to get times with details by swimmer
func (h *TimeHandler) getTimesWithDetailsBySwimmer(swimmerID int) ([]models.SwimTimeWithDetails, error) {
	query := `
		SELECT 
			st.id, st.swimmer_id, st.event_id, st.meet_id, st.time_ms, st.notes, st.recorded_at,
			s.name as swimmer_name,
			e.name as event_name,
			str.name as stroke_name,
			e.distance,
			m.name as meet_name
		FROM swim_times st
		JOIN swimmers s ON st.swimmer_id = s.id
		JOIN events e ON st.event_id = e.id
		JOIN strokes str ON e.stroke_id = str.id
		LEFT JOIN meets m ON st.meet_id = m.id
		WHERE st.swimmer_id = ?
		ORDER BY st.recorded_at DESC
	`

	rows, err := h.db.Query(query, swimmerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var times []models.SwimTimeWithDetails
	for rows.Next() {
		var timeDetails models.SwimTimeWithDetails
		var meetName sql.NullString

		err := rows.Scan(
			&timeDetails.ID, &timeDetails.SwimmerID, &timeDetails.EventID,
			&timeDetails.MeetID, &timeDetails.TimeMs, &timeDetails.Notes, &timeDetails.RecordedAt,
			&timeDetails.SwimmerName, &timeDetails.EventName, &timeDetails.StrokeName,
			&timeDetails.Distance, &meetName,
		)

		if err != nil {
			return nil, err
		}

		if meetName.Valid {
			timeDetails.MeetName = &meetName.String
		}

		// Format the time
		timeDetails.FormattedTime = timeDetails.SwimTime.FormatTime()

		times = append(times, timeDetails)
	}

	return times, nil
}

// Helper function to get all times with details
func (h *TimeHandler) getAllTimesWithDetails() ([]models.SwimTimeWithDetails, error) {
	query := `
		SELECT 
			st.id, st.swimmer_id, st.event_id, st.meet_id, st.time_ms, st.notes, st.recorded_at,
			s.name as swimmer_name,
			e.name as event_name,
			str.name as stroke_name,
			e.distance,
			m.name as meet_name
		FROM swim_times st
		JOIN swimmers s ON st.swimmer_id = s.id
		JOIN events e ON st.event_id = e.id
		JOIN strokes str ON e.stroke_id = str.id
		LEFT JOIN meets m ON st.meet_id = m.id
		ORDER BY st.recorded_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var times []models.SwimTimeWithDetails
	for rows.Next() {
		var timeDetails models.SwimTimeWithDetails
		var meetName sql.NullString

		err := rows.Scan(
			&timeDetails.ID, &timeDetails.SwimmerID, &timeDetails.EventID,
			&timeDetails.MeetID, &timeDetails.TimeMs, &timeDetails.Notes, &timeDetails.RecordedAt,
			&timeDetails.SwimmerName, &timeDetails.EventName, &timeDetails.StrokeName,
			&timeDetails.Distance, &meetName,
		)

		if err != nil {
			return nil, err
		}

		if meetName.Valid {
			timeDetails.MeetName = &meetName.String
		}

		// Format the time
		timeDetails.FormattedTime = timeDetails.SwimTime.FormatTime()

		times = append(times, timeDetails)
	}

	return times, nil
}
