// Command producer is the ingest half of the split Redpanda deployment: it
// streams the Wikimedia recent-changes feed and publishes each update to the
// updates topic.
package main

import (
	"wiki_updates/configuration"
	"wiki_updates/data"
	"wiki_updates/models"
	"wiki_updates/wiki"
)

func main() {
	config := configuration.GetConfig()
	producer := data.NewProducer(&config)
	defer producer.Close()
	wiki.ConsumeWikipediaChanges(&config, func(update models.Update) {
		producer.Produce(update)
	})
}
