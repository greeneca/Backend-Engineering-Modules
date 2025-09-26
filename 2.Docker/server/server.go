package server

import (
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/statstics"

	"github.com/gin-gonic/gin"
)


func Server(stats *statstics.Statstics, config configuration.Config) {
	router := gin.Default()
	router.GET("/stats", getStatsHandler(stats))
	router.Run(":" + config.ServerPort())
}

func getStatsHandler(stats *statstics.Statstics) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.JSON(http.StatusOK, generateStatsJson(stats))
	}
	return fn
}

func generateStatsJson(stats *statstics.Statstics) gin.H {
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
