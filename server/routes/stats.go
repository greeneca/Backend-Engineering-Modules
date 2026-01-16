package routes

import (
	"fmt"
	"net/http"
	"wiki_updates/models"

	"github.com/gin-gonic/gin"
)


func getStatsHandler(channel *chan models.Message) gin.HandlerFunc {
	statsFetcher := func() *models.Statistics {
		*channel <- models.Message{
			Type: "get_stats",
		}
		msg := <-*channel
		return msg.Statistics
	}
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, generateStatsJson(statsFetcher))
	}
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
