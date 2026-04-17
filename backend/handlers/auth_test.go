package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/laplogger/laplogger/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewAuthHandler(t *testing.T) {
	cfg := &config.Config{
		GoogleClientID:     "test-client-id",
		GoogleClientSecret: "test-client-secret",
		GoogleRedirectURI:  "http://localhost/callback",
		JWTSecret:          "test-secret",
	}

	h := NewAuthHandler(cfg, nil)
	if h == nil {
		t.Fatal("NewAuthHandler returned nil")
	}
	if h.oauthCfg.ClientID != "test-client-id" {
		t.Errorf("ClientID = %q; want %q", h.oauthCfg.ClientID, "test-client-id")
	}
	if h.oauthCfg.RedirectURL != "http://localhost/callback" {
		t.Errorf("RedirectURL = %q; want %q", h.oauthCfg.RedirectURL, "http://localhost/callback")
	}
	if len(h.oauthCfg.Scopes) != 3 {
		t.Errorf("Scopes count = %d; want 3", len(h.oauthCfg.Scopes))
	}
}

func TestGoogleLogin_Redirects(t *testing.T) {
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

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("status = %d; want %d", w.Code, http.StatusTemporaryRedirect)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("missing Location header on redirect")
	}

	if !strings.Contains(location, "accounts.google.com") {
		t.Errorf("Location = %q; should contain accounts.google.com", location)
	}
	if !strings.Contains(location, "client_id=test-client-id") {
		t.Errorf("Location = %q; should contain client_id=test-client-id", location)
	}
}

func TestGoogleCallback_MissingCode(t *testing.T) {
	cfg := &config.Config{
		GoogleClientID: "test-client-id",
		JWTSecret:      "test-secret",
	}

	h := NewAuthHandler(cfg, nil)

	r := gin.New()
	r.GET("/api/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetMe_NotAuthenticated(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	h := NewAuthHandler(cfg, nil)

	r := gin.New()
	r.GET("/api/me", h.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestIssueJWT_ValidToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	h := NewAuthHandler(cfg, nil)

	tokenStr, err := h.issueJWT("user-123", "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("issueJWT error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("issueJWT returned empty token")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	if err != nil {
		t.Fatalf("parse JWT error: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("could not cast claims to MapClaims")
	}

	if sub, _ := claims.GetSubject(); sub != "user-123" {
		t.Errorf("sub = %q; want %q", sub, "user-123")
	}
	if email, ok := claims["email"].(string); !ok || email != "test@example.com" {
		t.Errorf("email = %v; want test@example.com", claims["email"])
	}
	if name, ok := claims["name"].(string); !ok || name != "Test User" {
		t.Errorf("name = %v; want Test User", claims["name"])
	}
	if _, ok := claims["exp"]; !ok {
		t.Error("missing exp claim")
	}
	if _, ok := claims["iat"]; !ok {
		t.Error("missing iat claim")
	}
}

func TestIssueJWT_UsesHS256(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-key"}
	h := NewAuthHandler(cfg, nil)

	tokenStr, err := h.issueJWT("u1", "e@e.com", "N")
	if err != nil {
		t.Fatalf("issueJWT error: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte("test-key"), nil
	})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !token.Valid {
		t.Error("token should be valid")
	}
}
