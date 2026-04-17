package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/laplogger/laplogger/config"
	"github.com/laplogger/laplogger/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthHandler handles Google OAuth login and JWT issuance.
type AuthHandler struct {
	oauthCfg  *oauth2.Config
	jwtSecret []byte
	controlDB *database.ControlDB
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(cfg *config.Config, controlDB *database.ControlDB) *AuthHandler {
	return &AuthHandler{
		oauthCfg: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURI,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		jwtSecret: []byte(cfg.JWTSecret),
		controlDB: controlDB,
	}
}

// GoogleLogin redirects the user to Google's OAuth consent screen.
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// googleUserInfo represents the subset of fields returned by Google's userinfo endpoint.
type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GoogleCallback handles the OAuth callback, upserts the user, and issues a JWT.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code parameter"})
		return
	}

	// Exchange authorization code for token
	tok, err := h.oauthCfg.Exchange(c.Request.Context(), code)
	if err != nil {
		log.Printf("OAuth exchange error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "oauth exchange failed"})
		return
	}

	// Fetch user info from Google
	client := h.oauthCfg.Client(c.Request.Context(), tok)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		log.Printf("Userinfo fetch error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("Userinfo decode error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode user info"})
		return
	}

	// Upsert user in control DB
	user, err := h.controlDB.UpsertUser(info.Sub, info.Email, info.Name)
	if err != nil {
		log.Printf("User upsert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	// Issue JWT
	jwtToken, err := h.issueJWT(user.ID, user.Email, user.Name)
	if err != nil {
		log.Printf("JWT issue error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	// Redirect to frontend with token as query param
	frontendURL := fmt.Sprintf("/?token=%s", jwtToken)
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

// GetMe returns the current authenticated user's info.
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.controlDB.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) issueJWT(userID, email, name string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"name":  name,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}
