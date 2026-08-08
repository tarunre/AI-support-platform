package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a gin middleware that enforces Cross-Origin Resource Sharing policy.
// Allowed origins are read from the CORS_ALLOWED_ORIGINS environment variable
// (comma-separated). Falls back to localhost for local development.
func CORS() gin.HandlerFunc {
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var origins []string
	if originsEnv != "" {
		for _, o := range strings.Split(originsEnv, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}
	if len(origins) == 0 {
		origins = []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			"http://localhost:5176",
			"http://localhost:3000",
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
