package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/database"
)

func TestNewResultsHandler(t *testing.T) {
	h := NewResultsHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil ResultsHandler")
	}
}

func TestNewResultsHandler_Fields(t *testing.T) {
	ctrl := &database.ControlDB{}
	tm := &database.TenantManager{}
	h := NewResultsHandler(ctrl, tm)
	if h.controlDB != ctrl {
		t.Error("expected controlDB to be set")
	}
	if h.tenantMgr != tm {
		t.Error("expected tenantMgr to be set")
	}
}

func TestGetPublicResults_InvalidMeetID(t *testing.T) {
	h := NewResultsHandler(nil, nil)

	r := gin.New()
	r.GET("/api/public/:leagueSlug/meets/:meetId", h.GetPublicResults)

	req := httptest.NewRequest(http.MethodGet, "/api/public/test-league/meets/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "invalid meet id") {
		t.Errorf("body = %s; want 'invalid meet id'", w.Body.String())
	}
}

func TestGetPublicResults_MissingSlug(t *testing.T) {
	h := NewResultsHandler(nil, nil)

	// Gin won't route an empty param segment, so test with a direct handler call
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "leagueSlug", Value: ""},
		{Key: "meetId", Value: "550e8400-e29b-41d4-a716-446655440000"},
	}
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	h.GetPublicResults(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "missing league slug") {
		t.Errorf("body = %s; want 'missing league slug'", w.Body.String())
	}
}

func TestGetPublicResults_ValidUUIDAccepted(t *testing.T) {
	// Verify that a valid UUID passes the validation gate.
	// It will fail at the DB layer (nil controlDB) but should not return 400.
	// We can't test further without a real DB, but this confirms validation ordering.
	h := NewResultsHandler(nil, nil)

	r := gin.New()
	r.GET("/api/public/:leagueSlug/meets/:meetId", h.GetPublicResults)

	// Use a valid slug and UUID — should pass validation and fail at DB (not panic from validation)
	req := httptest.NewRequest(http.MethodGet,
		"/api/public/test-league/meets/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	// Expect a panic from nil DB access — this confirms validation passed
	defer func() {
		// Expected: nil pointer dereference on controlDB.GetLeagueBySlug
		// This proves the UUID validation passed (didn't return 400)
		_ = recover()
	}()
	r.ServeHTTP(w, req)

	// If we get here without panic, the status should NOT be 400
	if w.Code == http.StatusBadRequest {
		t.Error("valid UUID should not return 400")
	}
}

