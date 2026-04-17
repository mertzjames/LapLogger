package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/config"
)

// === SEC-1: OAuth Security Tests ===

// TestSecurity_OAuth_HardcodedStateParam documents that the OAuth state
// parameter is hardcoded to "state", making CSRF protection ineffective.
func TestSecurity_OAuth_HardcodedStateParam(t *testing.T) {
	cfg := &config.Config{
		GoogleClientID:     "test-client-id",
		GoogleClientSecret: "test-client-secret",
		GoogleRedirectURI:  "http://localhost/callback",
		JWTSecret:          "test-secret",
	}
	h := NewAuthHandler(cfg, nil)

	r := gin.New()
	r.GET("/api/auth/google", h.GoogleLogin)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	if !strings.Contains(location, "state=state") {
		t.Fatalf("expected state=state in redirect URL, got: %s", location)
	}

	// FINDING: The state parameter is hardcoded to "state".
	t.Log("SEC-FINDING: OAuth state parameter is hardcoded ('state'), CSRF protection is ineffective")
}

// TestSecurity_OAuth_CallbackNoStateValidation documents that the callback
// handler does not validate the state parameter returned by Google.
func TestSecurity_OAuth_CallbackNoStateValidation(t *testing.T) {
	cfg := &config.Config{
		GoogleClientID:     "test-client-id",
		GoogleClientSecret: "test-client-secret",
		GoogleRedirectURI:  "http://localhost/callback",
		JWTSecret:          "test-secret",
	}
	h := NewAuthHandler(cfg, nil)

	r := gin.New()
	r.GET("/api/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=test&state=evil", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusBadRequest {
		t.Log("Good: callback validates state parameter")
	} else {
		t.Log("SEC-FINDING: Callback does not validate OAuth state parameter (CSRF risk)")
	}
}

// TestSecurity_OAuth_TokenInURL documents JWT exposure via URL parameter.
func TestSecurity_OAuth_TokenInURL(t *testing.T) {
	t.Log("SEC-FINDING: JWT passed as URL query parameter (/?token=...) after OAuth callback")
	t.Log("  Risk: Token visible in browser history, HTTP Referer, proxy logs")
	t.Log("  Fix: Use HttpOnly cookie or POST with hidden form to deliver token")
}

// TestSecurity_OAuth_CallbackCodeInjection verifies that special characters
// in the code parameter don't cause unexpected behavior.
func TestSecurity_OAuth_CallbackCodeInjection(t *testing.T) {
	cfg := &config.Config{
		GoogleClientID:     "test-client-id",
		GoogleClientSecret: "test-client-secret",
		GoogleRedirectURI:  "http://localhost/callback",
		JWTSecret:          "test-secret",
	}
	h := NewAuthHandler(cfg, nil)

	r := gin.New()
	r.GET("/api/auth/google/callback", h.GoogleCallback)

	maliciousCodes := []string{
		"<script>alert(1)</script>",
		"'; DROP TABLE users;--",
		"code\r\nX-Injected: true",
		strings.Repeat("A", 10000),
	}

	for _, code := range maliciousCodes {
		label := code
		if len(label) > 30 {
			label = label[:30]
		}
		t.Run("code="+label, func(t *testing.T) {
			target := "/api/auth/google/callback?code=" + url.QueryEscape(code)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				t.Error("malicious code parameter should not result in 200 OK")
			}
		})
	}
}
