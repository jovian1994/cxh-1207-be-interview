package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/jwt"
	"net/http"
	"strings"
)

// LoginRequired 校验 JWT 的中间件
func LoginRequired(TokenVerify jwt.ITokenVerify) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		username, err := TokenVerify.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}
