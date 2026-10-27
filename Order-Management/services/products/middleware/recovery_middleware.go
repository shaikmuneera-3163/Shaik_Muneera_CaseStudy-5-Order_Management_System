package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware catches panics so one bad request does not crash the service.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID, _ := c.Get("request_id")
				log.Printf("panic recovered request_id=%v path=%s panic=%v\n%s", requestID, c.Request.URL.Path, recovered, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":     "internal server error",
					"requestId": requestID,
				})
			}
		}()
		c.Next()
	}
}
