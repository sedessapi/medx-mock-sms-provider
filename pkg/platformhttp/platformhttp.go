package platformhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderCorrelationID = "X-Correlation-ID"

func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderCorrelationID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("correlation_id", id)
		c.Header(HeaderCorrelationID, id)
		c.Next()
	}
}

func MountOps(r *gin.Engine, service string, openAPI map[string]any) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": service})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready", "service": service})
	})
	r.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, openAPI)
	})
	r.GET("/swagger.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, openAPI)
	})
}
