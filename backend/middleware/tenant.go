package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laplogger/laplogger/database"
)

// TenantRequired resolves the :leagueId URL param, verifies the authenticated
// user is a member, and attaches the tenant *sql.DB to the context as "tenantDB".
func TenantRequired(controlDB *database.ControlDB, tenantMgr *database.TenantManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Param("leagueId")
		if leagueID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing league id"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		// Verify the user is a member of this league
		isMember, err := controlDB.CheckMembership(userID.(string), leagueID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "membership check failed"})
			return
		}
		if !isMember {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not a member of this league"})
			return
		}

		// Look up the league to get the tenant DB name
		league, err := controlDB.GetLeagueByID(leagueID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve league"})
			return
		}
		if league == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "league not found"})
			return
		}

		// Get (or create) the tenant DB connection
		tenantDB, err := tenantMgr.GetDB(league.DBName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to league database"})
			return
		}

		c.Set("leagueID", leagueID)
		c.Set("tenantDB", tenantDB)
		c.Next()
	}
}

// GetTenantDB extracts the tenant *sql.DB from the Gin context.
func GetTenantDB(c *gin.Context) *sql.DB {
	db, _ := c.Get("tenantDB")
	return db.(*sql.DB)
}
