package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewTeamHandler(t *testing.T) {
	h := NewTeamHandler()
	if h == nil {
		t.Fatal("NewTeamHandler returned nil")
	}
}

func TestCreateTeam_MissingBody(t *testing.T) {
	h := NewTeamHandler()

	r := gin.New()
	r.POST("/teams", setTenantDB(nil), h.CreateTeam)

	req := httptest.NewRequest(http.MethodPost, "/teams", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateTeam_EmptyName(t *testing.T) {
	h := NewTeamHandler()

	r := gin.New()
	r.POST("/teams", setTenantDB(nil), h.CreateTeam)

	body := strings.NewReader(`{"name": "   "}`)
	req := httptest.NewRequest(http.MethodPost, "/teams", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateTeam_InvalidBody(t *testing.T) {
	h := NewTeamHandler()

	r := gin.New()
	r.PUT("/teams/:teamId", setTenantDB(nil), h.UpdateTeam)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPut, "/teams/some-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// setTenantDB is a test middleware that sets the tenantDB in the context.
func setTenantDB(db interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db != nil {
			c.Set("tenantDB", db)
		}
		c.Next()
	}
}
