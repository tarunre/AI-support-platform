package middleware

import (
	"net/http"

	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

// RequireRole returns middleware that permits access only to users whose role
// matches one of the provided values. SuperAdmin always passes.
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get("userRole")
		if !exists {
			utils.SendError(c, http.StatusForbidden, "Access denied")
			c.Abort()
			return
		}

		role := models.Role(roleVal.(string))

		// SuperAdmin has unrestricted access
		if role == models.RoleSuperAdmin {
			c.Next()
			return
		}

		if _, ok := allowed[role]; !ok {
			utils.SendError(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin restricts access to SuperAdmin only.
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin)
}
