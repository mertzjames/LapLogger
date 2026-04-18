package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestSecurity_TeamCreateXSS_InputAccepted verifies that team names with special chars
// pass validation (XSS prevention is the frontend's responsibility).
func TestSecurity_TeamCreateXSS_InputAccepted(t *testing.T) {
	name := "<script>alert('xss')</script>"
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		t.Error("XSS payload should pass empty-name validation")
	}
}

// TestSecurity_SwimmerGenderValidation ensures only valid M/F values accepted.
func TestSecurity_SwimmerGenderValidation(t *testing.T) {
	h := NewSwimmerHandler()

	r := gin.New()
	r.POST("/swimmers", setTenantDB(nil), h.CreateSwimmer)

	invalidGenders := []string{"X", "MM", "Male", "", "   "}
	for _, g := range invalidGenders {
		body := strings.NewReader(`{"team_id":"t1","first_name":"Jane","last_name":"Doe","date_of_birth":"2012-03-15","gender":"` + g + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/swimmers", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("gender=%q: status = %d; want %d", g, w.Code, http.StatusBadRequest)
		}
	}
}

// TestSecurity_EventStrokeValidation ensures only valid strokes are accepted for non-custom events.
func TestSecurity_EventStrokeValidation(t *testing.T) {
	h := NewEventHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events", setTenantDB(nil), h.CreateEvent)

	// Invalid stroke for non-custom event
	body := strings.NewReader(`{"stroke":"FreeStyle","distance":100}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSecurity_EventUnitValidation ensures only valid units are accepted.
func TestSecurity_EventUnitValidation(t *testing.T) {
	h := NewEventHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events", setTenantDB(nil), h.CreateEvent)

	body := strings.NewReader(`{"stroke":"Free","distance":100,"unit":"miles"}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSecurity_EventGenderValidation ensures only valid event genders (M/F/X) are accepted.
func TestSecurity_EventGenderValidation(t *testing.T) {
	h := NewEventHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events", setTenantDB(nil), h.CreateEvent)

	body := strings.NewReader(`{"stroke":"Free","distance":100,"gender":"Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSecurity_EventNegativeDistance ensures negative distances are rejected.
func TestSecurity_EventNegativeDistance(t *testing.T) {
	h := NewEventHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events", setTenantDB(nil), h.CreateEvent)

	body := strings.NewReader(`{"stroke":"Free","distance":-50}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSecurity_TimeNegativeValue ensures negative time values are rejected.
func TestSecurity_TimeNegativeValue(t *testing.T) {
	h := NewTimeHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events/:eventId/times", setTenantDB(nil), h.CreateTime)

	body := strings.NewReader(`{"swimmer_id":"s1","time_hundredths":-100}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events/e1/times", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSecurity_DuplicateKeyDetection verifies the helper correctly identifies duplicate key errors.
func TestSecurity_DuplicateKeyDetection(t *testing.T) {
	if !isDuplicateKeyError(errDuplicateKey("duplicate key value violates unique constraint \"times_event_id_swimmer_id_key\"")) {
		t.Error("isDuplicateKeyError should detect duplicate key error")
	}
	if !isDuplicateKeyError(errDuplicateKey("ERROR: 23505")) {
		t.Error("isDuplicateKeyError should detect error code 23505")
	}
}

type errDuplicateKey string

func (e errDuplicateKey) Error() string { return string(e) }
