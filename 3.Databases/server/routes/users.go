package routes

import (
	"net/http"
	"wiki_updates/models"

	"github.com/gin-gonic/gin"
)

func getSignupHandler(channel *chan models.Message) func(c *gin.Context) {
	userSaver := func(user models.User) error {
		*channel <- models.Message{
			Type: "save_user",
			User: user,
		}
		msg := <-*channel
		return msg.Error
	}
	return func(c *gin.Context) {
		response, err := signupUser(userSaver, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, response)
		}
	}
}

func signupUser(userSaver func(user models.User) error, c *gin.Context) (gin.H, error) {
	// Placeholder for user signup handler
	return gin.H{}, nil
}

func getLoginHandler(channel *chan models.Message) func(c *gin.Context) {
	userFetcher := func(email string) (models.User, error) {
		*channel <- models.Message{
			Type: "get_user",
			User: models.User{Email: email},
		}
		msg := <-*channel
		return msg.User, msg.Error
	}
	return func(c *gin.Context) {
		response, err := loginUser(userFetcher, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, response)
		}
	}
}

func loginUser(userFetcher func(string) (models.User, error), c *gin.Context) (gin.H, error) {
	// Placeholder for user login handler
	return gin.H{}, nil
}
