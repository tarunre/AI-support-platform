package middleware

import (
	"net/http"
	"strings"

	"github.com/ayush/supportiq/internal/config"
	jwtpkg "github.com/ayush/supportiq/internal/jwt"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Authenticate validates the Bearer token, loads the user from the database,
// and sets userID, userRole, tenantID, user, and tenant in the Gin context.
func Authenticate(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(c, http.StatusUnauthorized, "Authorization header missing or malformed")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtpkg.ValidateToken(tokenStr, cfg.JWTAccessSecret)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			utils.SendError(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		// Set user context
		c.Set("userID", user.ID)
		c.Set("userRole", string(user.Role))
		c.Set("team", user.Team)
		c.Set("user", &user)

		// SuperAdmin: tenant context is optional (all tenants accessible)
		if user.Role == models.RoleSuperAdmin {
			c.Set("tenantID", uuid.Nil)
			c.Next()
			return
		}

		// Load and validate tenant for regular users
		tenantID := user.TenantID
		if tenantID == uuid.Nil {
			utils.SendError(c, http.StatusUnauthorized, "User has no tenant assigned")
			c.Abort()
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, "id = ?", tenantID).Error; err != nil {
			utils.SendError(c, http.StatusUnauthorized, "Tenant not found")
			c.Abort()
			return
		}

		if tenant.Status == models.TenantStatusSuspended {
			utils.SendError(c, http.StatusForbidden, "Tenant account is suspended")
			c.Abort()
			return
		}

		if tenant.Status == models.TenantStatusDeleted {
			utils.SendError(c, http.StatusForbidden, "Tenant account no longer exists")
			c.Abort()
			return
		}

		c.Set("tenantID", tenantID)
		c.Next()
	}
}

// GetTenantID extracts the tenantID from the Gin context.
// Returns uuid.Nil for SuperAdmin (they have access to all tenants).
func GetTenantID(c *gin.Context) uuid.UUID {
	val, _ := c.Get("tenantID")
	id, _ := val.(uuid.UUID)
	return id
}
