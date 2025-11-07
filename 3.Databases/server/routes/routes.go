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
	authenticated.Use(getAuthMiddleware(config))
	authenticated.GET("/stats", getStatsHandler(channel))

	router.POST("/signup", getSignupHandler(channel))
	router.POST("/login", getLoginHandler(channel))

	return router
}

func getAuthMiddleware(config configuration.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		middlewares.Authenticate(c, config.JWTSecret())
	}
}


