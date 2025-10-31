package server

import (
	"fmt"
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/models"

	"github.com/gin-gonic/gin"
)


func Server(config configuration.Config, channel *chan models.Message) {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/stats")
	})
	router.GET("/stats", getStatsHandler(channel))
	router.Run(":" + config.ServerPort())
}

func getStatsHandler(channel *chan models.Message) gin.HandlerFunc {
	statsFetcher := func() *models.Statistics {
		*channel <- models.Message{
			Type: "get_stats",
		}
		msg := <-*channel
		return msg.Statistics
	}
	fn := func(c *gin.Context) {
		c.JSON(http.StatusOK, generateStatsJson(statsFetcher))
	}
	return fn
}

func generateStatsJson(statsFetcher func() *models.Statistics) gin.H {
	stats := statsFetcher()
	fmt.Println("Generating stats JSON:", stats.Messages)
	return gin.H{
		"messages": stats.Messages,
		"urls":     stats.Urls,
		"users":    stats.Bots + stats.NonBots,
		"bots":     stats.Bots,
		"non_bots": stats.NonBots,
	}
}
