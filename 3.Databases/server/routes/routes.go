package routes

import (
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/models"
	"wiki_updates/server/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter(config configuration.Config, channel *chan models.Message) *gin.Engine {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/stats")
	})
	authenticated := router.Group("/")
	authenticated.Use(getAuthMiddleware(config, channel))
	authenticated.GET("/stats", getStatsHandler(channel))

	router.POST("/signup", getSignupHandler(channel))
	router.POST("/login", getLoginHandler(channel, config))

	return router
}

func getAuthMiddleware(config configuration.Config, channel *chan models.Message) gin.HandlerFunc {
	userFetcher := func(email string) (models.User, error) {
		*channel <- models.Message{
			Type: "get_user",
			User: models.User{Email: email},
		}
		msg := <-*channel
		return msg.User, msg.Error
	}
	return func(c *gin.Context) {
		middlewares.Authenticate(c, config.JWTSecret(), userFetcher)
	}
}


