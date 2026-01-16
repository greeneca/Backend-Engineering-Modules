package server

import (
	"wiki_updates/configuration"
	"wiki_updates/models"
	"wiki_updates/server/routes"
)


func Server(config configuration.Config, channel *chan models.Message) {
	router := routes.SetupRouter(config, channel)
	router.Run(":" + config.ServerPort())
}

