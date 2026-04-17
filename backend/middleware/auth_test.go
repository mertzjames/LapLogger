package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var testJWTSecret = []byte("test-secret-key")

func createTestToken(secret []byte, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(secret)
	return tokenStr
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "missing authorization header" {
		t.Errorf("error = %q; want %q", body["error"], "missing authorization header")
	}
}

func TestAuthRequired_InvalidFormat_NoBearer(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "invalid authorization format" {
		t.Errorf("error = %q; want %q", body["error"], "invalid authorization format")
	}
}

func TestAuthRequired_InvalidFormat_TokenOnly(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "tokenonly")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequired_ExpiredToken(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
		"name":  "Test User",
		"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	})

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequired_WrongSecret(t *testing.T) {
	token := createTestToken([]byte("wrong-secret"), jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequired_MissingSubject(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequired_ValidToken(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
		"name":  "Test User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	var gotUserID, gotEmail, gotName string

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		uid, _ := c.Get("userID")
		gotUserID = uid.(string)
		email, _ := c.Get("email")
		gotEmail = email.(string)
		name, _ := c.Get("name")
		gotName = name.(string)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusOK)
	}
	if gotUserID != "user-123" {
		t.Errorf("userID = %q; want %q", gotUserID, "user-123")
	}
	if gotEmail != "test@example.com" {
		t.Errorf("email = %q; want %q", gotEmail, "test@example.com")
	}
	if gotName != "Test User" {
		t.Errorf("name = %q; want %q", gotName, "Test User")
	}
}

func TestAuthRequired_ValidToken_CaseInsensitiveBearer(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "BEARER "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d (case-insensitive BEARER should work)", w.Code, http.StatusOK)
	}
}

func TestAuthRequired_NonHMAC_SigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d (none algorithm should be rejected)", w.Code, http.StatusUnauthorized)
	}
}
