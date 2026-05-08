package main

import (
	"wiki_updates/configuration"
	"wiki_updates/data"
	"wiki_updates/models"
	"wiki_updates/server"
)


func main() {
	config := configuration.GetConfig()
	server_chan := make(chan models.Message)
	data.DataController(&config, server_chan)
	server.Server(&config, server_chan)
}

