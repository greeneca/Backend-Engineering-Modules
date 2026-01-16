package routes

import (
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/models"
	"wiki_updates/utils"

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
		signupUser(userSaver, c)
	}
}

func signupUser(userSaver func(user models.User) error, c *gin.Context) {
	user := models.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.PasswordHash = hashedPassword
	user.Password = ""
	err = userSaver(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User signed up successfully"})
}


func getLoginHandler(channel *chan models.Message, config configuration.Config) func(c *gin.Context) {
	userFetcher := func(email string) (models.User, error) {
		*channel <- models.Message{
			Type: "get_user",
			User: models.User{Email: email},
		}
		msg := <-*channel
		return msg.User, msg.Error
	}
	return func(c *gin.Context) {
		loginUser(userFetcher, config, c)
	}
}

func loginUser(userFetcher func(string) (models.User, error), config configuration.Config, c *gin.Context) {
	user := models.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}
	storedUser, err := userFetcher(user.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "details": err.Error()})
		return
	}
	if !utils.CompareHashAndPassword(storedUser.PasswordHash, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "details": "password mismatch"})
		return
	}
	token, err := utils.GenerateToken(storedUser.Email, config.JWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "email": storedUser.Email, "token": token})
}
