package middlewares

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(limiter *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(limiter, c.Writer, c.Request)
		if httpError != nil {
			c.JSON(httpError.StatusCode, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
