package middlewares

import (
	"net/http"
	"wiki_updates/models"
	"wiki_updates/utils"

	"github.com/gin-gonic/gin"
)

func Authenticate(c *gin.Context, secretKey string, userFetcher func(email string) (models.User, error)) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "No authorization token provided"})
		c.Abort()
		return
	}

	email, err := utils.ValidateToken(token, secretKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token", "error": err.Error()})
		c.Abort()
		return
	}

	user, err := userFetcher(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		c.Abort()
		return
	}

	c.Set("userEmail", user.Email)
	c.Next()
}
