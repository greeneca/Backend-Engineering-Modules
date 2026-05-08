package main

import (
	"wiki_updates/configuration"
	"wiki_updates/wiki"
)


func main() {
	config := configuration.GetConfig()
	wiki.ConsumeWikipediaChanges(&config)
}

