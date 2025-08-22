package main

import (
	"wiki_updates/server"
	"wiki_updates/statstics"
	"wiki_updates/wiki"
)


func main() {
	var stats = statstics.Statstics{
		Messages: 0,
		Urls:     make(map[string]bool),
		Bots:     make(map[string]bool),
		NonBots:  make(map[string]bool),
	}
	go server.Server(&stats)
	wiki.ConsumeWikipediaChanges(&stats)
}

