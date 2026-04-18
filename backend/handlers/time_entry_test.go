package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewTimeHandler(t *testing.T) {
	h := NewTimeHandler()
	if h == nil {
		t.Fatal("NewTimeHandler returned nil")
	}
}

func TestCreateTime_MissingBody(t *testing.T) {
	h := NewTimeHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events/:eventId/times", setTenantDB(nil), h.CreateTime)

	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events/e1/times", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateTime_NegativeTime(t *testing.T) {
	h := NewTimeHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events/:eventId/times", setTenantDB(nil), h.CreateTime)

	body := strings.NewReader(`{"swimmer_id":"s1","time_hundredths":-5}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events/e1/times", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateTime_ZeroTime(t *testing.T) {
	h := NewTimeHandler()

	r := gin.New()
	r.POST("/meets/:meetId/events/:eventId/times", setTenantDB(nil), h.CreateTime)

	body := strings.NewReader(`{"swimmer_id":"s1","time_hundredths":0}`)
	req := httptest.NewRequest(http.MethodPost, "/meets/m1/events/e1/times", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateTime_InvalidBody(t *testing.T) {
	h := NewTimeHandler()

	r := gin.New()
	r.PUT("/meets/:meetId/events/:eventId/times/:timeId", setTenantDB(nil), h.UpdateTime)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPut, "/meets/m1/events/e1/times/t1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"generic error", errors.New("some error"), false},
		{"duplicate key text", errors.New("duplicate key value violates unique constraint"), true},
		{"postgres code 23505", errors.New("ERROR: 23505 unique_violation"), true},
		{"unrelated error", errors.New("connection refused"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDuplicateKeyError(tt.err)
			if got != tt.want {
				t.Errorf("isDuplicateKeyError(%v) = %v; want %v", tt.err, got, tt.want)
			}
		})
	}
}
