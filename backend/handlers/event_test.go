package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewEventHandler(t *testing.T) {
	h := NewEventHandler()
	if h == nil {
		t.Fatal("NewEventHandler returned nil")
	}
}

func TestValidStrokes(t *testing.T) {
	expected := []string{"Free", "Back", "Breast", "Fly", "IM"}
	for _, s := range expected {
		if !validStrokes[s] {
			t.Errorf("validStrokes[%q] = false; want true", s)
		}
	}
	invalid := []string{"free", "BACK", "Freestyle", "Butterfly", ""}
	for _, s := range invalid {
		if validStrokes[s] {
			t.Errorf("validStrokes[%q] = true; want false", s)
		}
	}
}

func TestValidUnits(t *testing.T) {
	if !validUnits["yards"] {
		t.Error("validUnits[yards] = false; want true")
	}
	if !validUnits["meters"] {
		t.Error("validUnits[meters] = false; want true")
	}
	if validUnits["Yards"] {
		t.Error("validUnits[Yards] = true; want false")
	}
}

func TestValidEventGenders(t *testing.T) {
	for _, g := range []string{"M", "F", "X"} {
		if !validEventGenders[g] {
			t.Errorf("validEventGenders[%q] = false; want true", g)
		}
	}
	if validEventGenders["m"] {
		t.Error("validEventGenders[m] = true; want false")
	}
}

func TestCreateEvent_MissingBody(t *testing.T) {
	h := NewEventHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events", setTenantDB(nil), h.CreateEvent)

	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateEvent_InvalidBody(t *testing.T) {
	h := NewEventHandler()

	r := gin.New()
	r.PUT("/meets/:meetId/events/:eventId", setTenantDB(nil), h.UpdateEvent)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPut, "/meets/m1/events/e1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}
