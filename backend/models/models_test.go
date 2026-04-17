package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser_JSONSerialization(t *testing.T) {
	u := User{
		ID:        "abc-123",
		GoogleID:  "google-456",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded["id"] != "abc-123" {
		t.Errorf("id = %v; want abc-123", decoded["id"])
	}
	if decoded["email"] != "test@example.com" {
		t.Errorf("email = %v; want test@example.com", decoded["email"])
	}
}

func TestLeague_DBNameOmittedFromJSON(t *testing.T) {
	league := League{
		ID:        "league-1",
		Name:      "Test League",
		Slug:      "test-league",
		DBName:    "laplogger_league_abc",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(league)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, exists := decoded["db_name"]; exists {
		t.Error("db_name should be omitted from JSON (has json:\"-\" tag)")
	}
	if _, exists := decoded["DBName"]; exists {
		t.Error("DBName should be omitted from JSON")
	}
}

func TestLeague_JSONContainsExpectedFields(t *testing.T) {
	league := League{
		ID:   "l-1",
		Name: "Sharks",
		Slug: "sharks",
	}

	data, err := json.Marshal(league)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, field := range []string{"id", "name", "slug", "created_at"} {
		if _, exists := decoded[field]; !exists {
			t.Errorf("expected field %q missing from JSON", field)
		}
	}
}

func TestTimeEntry_TimeHundredths(t *testing.T) {
	entry := TimeEntry{
		ID:             "t-1",
		EventID:        "e-1",
		SwimmerID:      "s-1",
		TimeHundredths: 8345,
		IsExhibition:   false,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if th, ok := decoded["time_hundredths"].(float64); !ok || int(th) != 8345 {
		t.Errorf("time_hundredths = %v; want 8345", decoded["time_hundredths"])
	}
}

func TestEvent_CustomEvent(t *testing.T) {
	event := Event{
		ID:         "e-1",
		MeetID:     "m-1",
		Stroke:     "Free",
		Distance:   100,
		Unit:       "yards",
		Gender:     "M",
		AgeGroup:   "11-12",
		IsCustom:   true,
		CustomName: "Sprint Challenge",
		SortOrder:  1,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded["is_custom"] != true {
		t.Errorf("is_custom = %v; want true", decoded["is_custom"])
	}
	if decoded["custom_name"] != "Sprint Challenge" {
		t.Errorf("custom_name = %v; want Sprint Challenge", decoded["custom_name"])
	}
}

func TestEvent_StandardEvent_OmitsCustomName(t *testing.T) {
	event := Event{
		ID:       "e-2",
		MeetID:   "m-1",
		Stroke:   "Back",
		Distance: 50,
		Unit:     "meters",
		Gender:   "F",
		AgeGroup: "13-14",
		IsCustom: false,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, exists := decoded["custom_name"]; exists {
		t.Error("custom_name should be omitted for standard events (omitempty)")
	}
}

func TestSwimmer_GenderValues(t *testing.T) {
	for _, gender := range []string{"M", "F"} {
		s := Swimmer{
			ID:          "s-1",
			TeamID:      "t-1",
			FirstName:   "Test",
			LastName:    "Swimmer",
			DateOfBirth: "2012-03-15",
			Gender:      gender,
		}

		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal error for gender %q: %v", gender, err)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if decoded["gender"] != gender {
			t.Errorf("gender = %v; want %q", decoded["gender"], gender)
		}
	}
}
