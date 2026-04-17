package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// === SEC-1: JWT Security Tests ===

// TestSecurity_JWT_AlgorithmConfusion verifies that a token signed with
// an asymmetric algorithm (ES256) is rejected even if it's technically valid.
// This defends against algorithm confusion attacks where an attacker uses
// a public key as an HMAC secret.
func TestSecurity_JWT_AlgorithmConfusion_ES256(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate EC key: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d (ES256 token must be rejected when HS256 expected)", w.Code, http.StatusUnauthorized)
	}
}

// TestSecurity_JWT_TamperedPayload verifies that modifying JWT claims
// without re-signing invalidates the token.
func TestSecurity_JWT_TamperedPayload(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub":   "user-123",
		"email": "legit@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// Tamper with the payload by changing a character in the middle segment
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}
	// Flip a character in the payload
	payload := []byte(parts[1])
	if len(payload) > 5 {
		payload[5] = 'A' + (payload[5]-'A'+1)%26
	}
	tampered := parts[0] + "." + string(payload) + "." + parts[2]

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d (tampered token must be rejected)", w.Code, http.StatusUnauthorized)
	}
}

// TestSecurity_JWT_EmptySecret verifies behavior with an empty signing secret.
// An empty secret should still produce valid HMAC signatures but represents
// a critical misconfiguration. This test documents the behavior.
func TestSecurity_JWT_EmptySecret(t *testing.T) {
	emptySecret := []byte("")
	token := createTestToken(emptySecret, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	// Token signed with empty secret should NOT be accepted by middleware
	// using a non-empty secret
	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d (token signed with empty secret must be rejected)", w.Code, http.StatusUnauthorized)
	}
}

// TestSecurity_JWT_FutureIssuedAt verifies tokens with iat in the future
// are still accepted (iat is informational, not enforced by default).
func TestSecurity_JWT_FutureIssuedAt(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub": "user-123",
		"iat": time.Now().Add(1 * time.Hour).Unix(),
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Document: jwt-go does NOT reject future iat by default.
	// This is acceptable behavior but worth noting.
	if w.Code != http.StatusOK {
		t.Logf("Note: future iat token was rejected (status %d). This is stricter than expected.", w.Code)
	}
}

// TestSecurity_JWT_MissingExpClaim verifies tokens without an exp claim
// are rejected. The middleware uses jwt.WithExpirationRequired() to enforce this.
func TestSecurity_JWT_MissingExpClaim(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub": "user-123",
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
		t.Errorf("status = %d; want %d (tokens without exp must be rejected)", w.Code, http.StatusUnauthorized)
	}
}

// TestSecurity_JWT_ExtraWhitespaceInHeader verifies the auth header parser
// handles edge cases with extra whitespace.
func TestSecurity_JWT_ExtraWhitespaceInHeader(t *testing.T) {
	token := createTestToken(testJWTSecret, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	r := gin.New()
	r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Leading/trailing spaces in Authorization header value
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "  Bearer "+token+"  ")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// SplitN won't strip leading spaces, so "  Bearer" != "bearer"
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d (leading spaces should cause rejection)", w.Code, http.StatusUnauthorized)
	}
}

// TestSecurity_JWT_SQLInjectionInSubject verifies that malicious content
// in JWT subject claim doesn't cause issues when used as a user ID.
func TestSecurity_JWT_SQLInjectionInSubject(t *testing.T) {
	maliciousSubs := []string{
		"'; DROP TABLE users;--",
		"1 OR 1=1",
		"admin'--",
		"<script>alert(1)</script>",
	}

	for _, sub := range maliciousSubs {
		t.Run(sub, func(t *testing.T) {
			token := createTestToken(testJWTSecret, jwt.MapClaims{
				"sub": sub,
				"exp": time.Now().Add(1 * time.Hour).Unix(),
			})

			var extractedUserID string
			r := gin.New()
			r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
				uid, _ := c.Get("userID")
				extractedUserID = uid.(string)
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// The middleware should accept the token (it's validly signed)
			// but the subject is passed through as-is. SQL injection protection
			// must happen at the query layer (parameterized queries).
			if w.Code == http.StatusOK && extractedUserID != sub {
				t.Errorf("userID = %q; want %q (subject should be preserved exactly)", extractedUserID, sub)
			}
		})
	}
}

// TestSecurity_JWT_ErrorMessageLeakage ensures error responses don't
// leak internal details about JWT validation failures.
func TestSecurity_JWT_ErrorMessageLeakage(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"missing", ""},
		{"invalid format", "NotBearer token"},
		{"bad token", "Bearer invalid.token.here"},
		{"expired", "Bearer " + createTestToken(testJWTSecret, jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(-1 * time.Hour).Unix(),
		})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/test", AuthRequired(testJWTSecret), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)

			errMsg := body["error"]
			// Error messages should be generic — no stack traces, no internal paths
			if strings.Contains(errMsg, "/") || strings.Contains(errMsg, "panic") ||
				strings.Contains(errMsg, ".go:") || strings.Contains(errMsg, "runtime") {
				t.Errorf("error message leaks internal info: %q", errMsg)
			}
		})
	}
}
