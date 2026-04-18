package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewMeetHandler(t *testing.T) {
	h := NewMeetHandler()
	if h == nil {
		t.Fatal("NewMeetHandler returned nil")
	}
}

func TestCreateMeet_MissingBody(t *testing.T) {
	h := NewMeetHandler()

	r := gin.New()
	r.POST("/meets", setTenantDB(nil), h.CreateMeet)

	req := httptest.NewRequest(http.MethodPost, "/meets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateMeet_EmptyName(t *testing.T) {
	h := NewMeetHandler()

	r := gin.New()
	r.POST("/meets", setTenantDB(nil), h.CreateMeet)

	body := strings.NewReader(`{"name": "   ", "meet_date": "2026-04-15"}`)
	req := httptest.NewRequest(http.MethodPost, "/meets", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateMeet_InvalidBody(t *testing.T) {
	h := NewMeetHandler()

	r := gin.New()
	r.PUT("/meets/:meetId", setTenantDB(nil), h.UpdateMeet)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPut, "/meets/some-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}
