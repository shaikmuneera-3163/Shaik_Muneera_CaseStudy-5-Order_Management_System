package correlation

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Set(c *gin.Context) {
	id := c.GetHeader("X-Correlation-ID")

	if id == "" {
		id = uuid.NewString()
	}

	c.Set("correlation_id", id)
	c.Header("X-Correlation-ID", id)
}

func Get(c *gin.Context) string {
	value, exists := c.Get("correlation_id")
	if !exists {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}

	return id
}
