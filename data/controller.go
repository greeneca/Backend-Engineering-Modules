package data

//go:generate mockgen -source=controller.go -destination=mock/datasource.go

import (
	"fmt"
	"wiki_updates/configuration"
	"wiki_updates/models"
)

type DataSource interface {
	Initialize(configuration.Config)
	SaveUpdate(models.Update) error
	GetStatistics() (*models.Statistics, error)
	SaveUser(*models.User) error
	GetUserByEmail(string) (*models.User, error)
}

func DataController(config configuration.Config, wiki_chan *chan models.Message, server_chan *chan models.Message) {
	dataSource := getDataSource(config)
	dataSource.Initialize(config)
	monitorChannels(wiki_chan, server_chan, dataSource)
}
func getDataSource(config configuration.Config) DataSource {
	var dataSource DataSource
	if config.DataStorage() == "cassandra" {
		fmt.Println("Using Cassandra as data source")
		dataSource = &Cassandra{}
	} else {
		// Fallback to in-memory or other data source
		fmt.Println("Using In-Memory as data source")
		dataSource = &InMemory{}
	}
	return dataSource
}
func monitorChannels(wiki_chan *chan models.Message, server_chan *chan models.Message, dataSource DataSource) {
	for {
		select {
		case msg := <-*server_chan:
			switch msg.Type {
			case "get_stats":
				stats, err := dataSource.GetStatistics()
				if err != nil {
					fmt.Println("Error getting initial statistics:", err)
					stats = &models.Statistics{}
				}
				*server_chan <- models.Message{
					Type:       "stats_response",
					Statistics: stats,
				}
			case "save_user":
				err := dataSource.SaveUser(&msg.User)
				*server_chan <- models.Message{
					Type:  "save_user_response",
					Error: err,
				}
			case "get_user":
				user, err := dataSource.GetUserByEmail(msg.User.Email)
				if err != nil {
					*server_chan <- models.Message{
						Type:  "get_user_response",
						Error: err,
					}
				} else {
					*server_chan <- models.Message{
						Type: "get_user_response",
						User: *user,
					}
				}
			}
		case msg := <-*wiki_chan:
			if msg.Type == "save_data" {
				err := dataSource.SaveUpdate(msg.Update)
				if err != nil {
					fmt.Println("Error saving update:", err)
				}
			}
		}
	}
}
