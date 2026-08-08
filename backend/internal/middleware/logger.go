package middleware

import (
	"time"

	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogger returns a gin middleware that logs every incoming HTTP request
// using the application structured logger.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		utils.Logger.WithFields(logrus.Fields{
			"status":  c.Writer.Status(),
			"method":  method,
			"path":    path,
			"latency": time.Since(start).String(),
			"ip":      c.ClientIP(),
		}).Info("request")
	}
}
