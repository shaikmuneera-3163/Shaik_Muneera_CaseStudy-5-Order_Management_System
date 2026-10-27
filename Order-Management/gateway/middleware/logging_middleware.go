package middleware

import (
	"log"
	"time"

	"order-management/shared/correlation"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		correlation.Set(c)
		c.Next()
		requestID, _ := c.Get("request_id")

		log.Printf("service=gateway method=%s path=%s status=%d duration=%s correlation_id=%s request_id=%v",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(),
			time.Since(start), correlation.Get(c), requestID)
	}
}
