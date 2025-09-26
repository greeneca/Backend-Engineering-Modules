package main

import (
	"wiki_updates/configuration"
	"wiki_updates/server"
	"wiki_updates/statistics"
	"wiki_updates/wiki"
)


func main() {
	config := configuration.GetConfig()
	var stats = statistics.Statistics{
		Messages: 0,
		Urls:     make(map[string]bool),
		Bots:     make(map[string]bool),
		NonBots:  make(map[string]bool),
	}
	go server.Server(&stats, config)
	wiki.ConsumeWikipediaChanges(&stats, config)
}

