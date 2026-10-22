// Command consumer is the processing/serving half of the split Redpanda
// deployment: it consumes the updates topic into an in-memory materialized view
// (with snapshot recovery) and serves the HTTP API.
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
	go data.RunConsumer(&config, server_chan)
	server.Server(&config, server_chan)
}
