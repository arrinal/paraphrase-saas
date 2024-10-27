package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				log.Printf("Panic recovered: %v\nStack trace:\n%s", err, debug.Stack())

				// Generate a unique request ID
				requestID := c.GetString("RequestID")

				// Return a safe error response
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:     "An unexpected error occurred",
					Code:      "INTERNAL_SERVER_ERROR",
					RequestID: requestID,
				})

				c.Abort()
			}
		}()

		c.Next()

		// Handle any errors that were added to the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Log the error
			log.Printf("Error handling request: %v", err)

			// Return an appropriate error response
			status := http.StatusInternalServerError
			if err.IsType(gin.ErrorTypeBind) {
				status = http.StatusBadRequest
			}

			c.JSON(status, ErrorResponse{
				Error: err.Error(),
				Code:  string(err.Type), // Convert to string instead of calling Error()
			})
		}
	}
}
