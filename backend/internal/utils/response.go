package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// successResponse is the standard envelope for successful API responses.
type successResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// errorResponse is the standard envelope for error API responses.
type errorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// SendSuccess writes a JSON success response with the given status code.
func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, successResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// SendError writes a JSON error response with the given status code.
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, errorResponse{
		Status:  "error",
		Message: message,
	})
}

// SendInternalError logs the underlying error and returns a generic 500 to the
// client so internal details (DB schema, stack traces) are never exposed.
func SendInternalError(c *gin.Context, err error) {
	Logger.WithError(err).Error("internal server error")
	c.JSON(http.StatusInternalServerError, errorResponse{
		Status:  "error",
		Message: "internal server error",
	})
}

// SafeFilename strips characters that can break Content-Disposition headers
// (quotes, backslashes, newlines) to prevent header injection.
func SafeFilename(name string) string {
	r := strings.NewReplacer(`"`, "", `\`, "", "\n", "", "\r", "")
	return r.Replace(name)
}
