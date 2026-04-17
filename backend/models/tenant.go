package models

import "time"

// Team represents a swim team within a league.
type Team struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ShortName string    `json:"short_name" db:"short_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Swimmer represents an individual swimmer on a team.
type Swimmer struct {
	ID          string    `json:"id" db:"id"`
	TeamID      string    `json:"team_id" db:"team_id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	DateOfBirth string    `json:"date_of_birth" db:"date_of_birth"`
	Gender      string    `json:"gender" db:"gender"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Meet represents a swim meet.
type Meet struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Location  string    `json:"location" db:"location"`
	MeetDate  string    `json:"meet_date" db:"meet_date"`
	IsPublic  bool      `json:"is_public" db:"is_public"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Event represents a specific event at a meet (e.g., Boys 100 Free, 11-12).
type Event struct {
	ID         string    `json:"id" db:"id"`
	MeetID     string    `json:"meet_id" db:"meet_id"`
	Stroke     string    `json:"stroke" db:"stroke"`
	Distance   int       `json:"distance" db:"distance"`
	Unit       string    `json:"unit" db:"unit"`
	Gender     string    `json:"gender" db:"gender"`
	AgeGroup   string    `json:"age_group" db:"age_group"`
	IsCustom   bool      `json:"is_custom" db:"is_custom"`
	CustomName string    `json:"custom_name,omitempty" db:"custom_name"`
	SortOrder  int       `json:"sort_order" db:"sort_order"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// TimeEntry represents a swimmer's recorded time in an event.
type TimeEntry struct {
	ID             string    `json:"id" db:"id"`
	EventID        string    `json:"event_id" db:"event_id"`
	SwimmerID      string    `json:"swimmer_id" db:"swimmer_id"`
	TimeHundredths int       `json:"time_hundredths" db:"time_hundredths"`
	IsExhibition   bool      `json:"is_exhibition" db:"is_exhibition"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
