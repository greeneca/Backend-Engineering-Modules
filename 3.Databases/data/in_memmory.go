package data

import (
	"wiki_updates/configuration"
	"wiki_updates/models"
)

type InMemory struct {
	statistics models.Statistics
	urls 	map[string]bool
	bots 	map[string]bool
	nonBots map[string]bool
}

func (im *InMemory) Initialize(config configuration.Config) {
	im.statistics = models.Statistics{
		Messages: 0,
		Urls:     0,
		Bots:     0,
		NonBots:  0,
	}
}

func (im *InMemory) SaveUpdate(update models.Update) error {
	im.statistics.Messages++
	im.urls[update.Uri] = true
	if update.Bot {
		im.bots[update.User] = true
	} else {
		im.nonBots[update.User] = true
	}
	return nil
}

func (im *InMemory) GetStatistics() (*models.Statistics, error) {
	im.statistics.Urls = len(im.urls)
	im.statistics.Bots = len(im.bots)
	im.statistics.NonBots = len(im.nonBots)
	return &im.statistics, nil
}
