package server

import (
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/statistics"

	"github.com/gin-gonic/gin"
)


func Server(stats *statistics.Statistics, config configuration.Config) {
	router := gin.Default()
	router.GET("/stats", getStatsHandler(stats))
	router.Run(":" + config.ServerPort())
}

func getStatsHandler(stats *statistics.Statistics) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.JSON(http.StatusOK, generateStatsJson(stats))
	}
	return fn
}

func generateStatsJson(stats *statistics.Statistics) gin.H {
	stats.RLock()
	defer stats.RUnlock()
	return gin.H{
		"messages": stats.Messages,
		"urls":     len(stats.Urls),
		"users":    len(stats.Bots) + len(stats.NonBots),
		"bots":     len(stats.Bots),
		"non_bots": len(stats.NonBots),
	}
}
