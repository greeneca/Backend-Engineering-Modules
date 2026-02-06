package main

import (
	"wiki_updates/configuration"
	"wiki_updates/data"
	"wiki_updates/models"
	"wiki_updates/server"
	"wiki_updates/wiki"
)


func main() {
	config := configuration.GetConfig()
	wiki_chan := make(chan models.Message)
	server_chan := make(chan models.Message)
	go data.DataController(&config, &wiki_chan, &server_chan)
	go server.Server(&config, &server_chan)
	wiki.ConsumeWikipediaChanges(&config, wiki_chan)
}

