package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewSwimmerHandler(t *testing.T) {
	h := NewSwimmerHandler()
	if h == nil {
		t.Fatal("NewSwimmerHandler returned nil")
	}
}

func TestCreateSwimmer_MissingBody(t *testing.T) {
	h := NewSwimmerHandler()

	r := gin.New()
	r.POST("/swimmers", setTenantDB(nil), h.CreateSwimmer)

	req := httptest.NewRequest(http.MethodPost, "/swimmers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateSwimmer_EmptyNames(t *testing.T) {
	h := NewSwimmerHandler()

	r := gin.New()
	r.POST("/swimmers", setTenantDB(nil), h.CreateSwimmer)

	body := strings.NewReader(`{"team_id":"t1","first_name":"  ","last_name":"  ","date_of_birth":"2012-03-15","gender":"M"}`)
	req := httptest.NewRequest(http.MethodPost, "/swimmers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateSwimmer_InvalidGender(t *testing.T) {
	h := NewSwimmerHandler()

	r := gin.New()
	r.POST("/swimmers", setTenantDB(nil), h.CreateSwimmer)

	body := strings.NewReader(`{"team_id":"t1","first_name":"Jane","last_name":"Doe","date_of_birth":"2012-03-15","gender":"X"}`)
	req := httptest.NewRequest(http.MethodPost, "/swimmers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateSwimmer_InvalidBody(t *testing.T) {
	h := NewSwimmerHandler()

	r := gin.New()
	r.PUT("/swimmers/:swimmerId", setTenantDB(nil), h.UpdateSwimmer)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPut, "/swimmers/some-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}
