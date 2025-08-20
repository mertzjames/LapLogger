package models

import (
	"fmt"
	"time"
)

// User represents a user in the system with authentication
type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Don't include in JSON responses
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Swimmer represents a swimmer in the system
type Swimmer struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Meet represents a swimming meet/competition (single day)
type Meet struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Location    string    `json:"location" db:"location"`
	MeetDate    time.Time `json:"meet_date" db:"meet_date"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Event represents a specific swimming event (stroke + distance combination)
type Event struct {
	ID       int    `json:"id" db:"id"`
	StrokeID int    `json:"stroke_id" db:"stroke_id"`
	Distance int    `json:"distance" db:"distance"` // Distance in meters
	Name     string `json:"name" db:"name"`         // e.g., "50m Freestyle"
}

// EventWithDetails includes stroke information for display
type EventWithDetails struct {
	Event
	StrokeName string `json:"stroke_name"`
}

// Stroke represents a swimming stroke type
type Stroke struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// SwimTime represents a recorded swim time
type SwimTime struct {
	ID         int       `json:"id" db:"id"`
	SwimmerID  int       `json:"swimmer_id" db:"swimmer_id"`
	EventID    int       `json:"event_id" db:"event_id"`
	MeetID     *int      `json:"meet_id" db:"meet_id"` // Optional - can be practice time
	TimeMs     int       `json:"time_ms" db:"time_ms"` // Time in milliseconds
	Notes      string    `json:"notes" db:"notes"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
}

// SwimTimeWithDetails includes related information for display
type SwimTimeWithDetails struct {
	SwimTime
	SwimmerName   string  `json:"swimmer_name"`
	EventName     string  `json:"event_name"`
	StrokeName    string  `json:"stroke_name"`
	Distance      int     `json:"distance"`
	MeetName      *string `json:"meet_name"`
	FormattedTime string  `json:"formatted_time"`
}

// MeetEvent represents which events are scheduled for a specific meet
type MeetEvent struct {
	ID        int       `json:"id" db:"id"`
	MeetID    int       `json:"meet_id" db:"meet_id"`
	EventID   int       `json:"event_id" db:"event_id"`
	Session   string    `json:"session" db:"session"`     // "morning", "evening", "prelims", "finals", etc.
	EventNum  int       `json:"event_num" db:"event_num"` // Event number in the session
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MeetEventWithDetails includes full event and meet information
type MeetEventWithDetails struct {
	MeetEvent
	MeetName   string    `json:"meet_name"`
	MeetDate   time.Time `json:"meet_date"`
	EventName  string    `json:"event_name"`
	StrokeName string    `json:"stroke_name"`
	Distance   int       `json:"distance"`
}

// Request types for API
type CreateSwimmerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateMeetRequest struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	MeetDate    string `json:"meet_date"` // ISO date string
	Description string `json:"description"`
}

type CreateEventRequest struct {
	StrokeID int `json:"stroke_id"`
	Distance int `json:"distance"`
}

type CreateTimeRequest struct {
	SwimmerID int    `json:"swimmer_id"`
	EventID   int    `json:"event_id"`
	MeetID    *int   `json:"meet_id"` // Optional
	TimeMs    int    `json:"time_ms"`
	Notes     string `json:"notes"`
}

type CreateMeetEventRequest struct {
	MeetID   int    `json:"meet_id"`
	EventID  int    `json:"event_id"`
	Session  string `json:"session"`
	EventNum int    `json:"event_num"`
}

// FormatTime converts milliseconds to MM:SS.MS format
func (st *SwimTime) FormatTime() string {
	totalSeconds := st.TimeMs / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	milliseconds := st.TimeMs % 1000

	return fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, milliseconds)
}

// GenerateEventName creates a descriptive name for an event
func (e *Event) GenerateEventName(strokeName string) string {
	return fmt.Sprintf("%dm %s", e.Distance, strokeName)
}

// Authentication request/response types
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
