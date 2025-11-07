package middlewares

import (
	"net/http"
	"wiki_updates/utils"

	"github.com/gin-gonic/gin"
)

func Authenticate(c *gin.Context, secretKey string) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "No authorization token provided"})
		c.Abort()
		return
	}

	userID, err := utils.ValidateToken(token, secretKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token", "error": err.Error()})
		c.Abort()
		return
	}

	c.Set("userID", userID)
	c.Next()
}
