package data

import (
	"fmt"
	"wiki_updates/configuration"
	"wiki_updates/models"
)

type DataSource interface {
	Initialize(configuration.Config)
	SaveUpdate(models.Update) error
	GetStatistics() (*models.Statistics, error)
}

func DataController(config configuration.Config, wiki_chan *chan models.Message, server_chan *chan models.Message) {
	var dataSource DataSource
	if config.DataStorage() == "cassandra" {
		fmt.Println("Using Cassandra as data source")
		dataSource = &Cassandra{}
		dataSource.Initialize(config)
	} else {
		// Fallback to in-memory or other data source
		fmt.Println("Using In-Memory as data source")
		dataSource = &InMemory{}
		dataSource.Initialize(config)
	}
	for {
		select {
		case msg := <-*server_chan:
			if msg.Type == "get_stats" {
				stats, err := dataSource.GetStatistics()
				if err != nil {
					fmt.Println("Error getting initial statistics:", err)
					stats = &models.Statistics{}
				}
				*server_chan <- models.Message{
					Type:       "stats_response",
					Statistics: stats,
				}
			}
		case msg := <-*wiki_chan:
			if msg.Type == "save_data" {
				dataSource.SaveUpdate(msg.Update)
			}
		}
	}
}
