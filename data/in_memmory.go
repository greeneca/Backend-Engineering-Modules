package data

import (
	"errors"
	"wiki_updates/configuration"
	"wiki_updates/models"
)

type InMemory struct {
	statistics models.Statistics
	urls       map[string]bool
	bots       map[string]bool
	nonBots    map[string]bool
	users      map[string]*models.User
}

func (im *InMemory) Initialize(_ configuration.Config) {
	im.urls = make(map[string]bool)
	im.bots = make(map[string]bool)
	im.nonBots = make(map[string]bool)
	im.statistics = models.Statistics{
		Messages: 0,
		Urls:     0,
		Bots:     0,
		NonBots:  0,
	}
	im.users = make(map[string]*models.User)
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

func (im *InMemory) GetUserByEmail(email string) (*models.User, error) {
	if _, exists := im.users[email]; !exists {
		return nil, errors.New("user not found")
	}
	return im.users[email], nil
}

func (im *InMemory) SaveUser(user *models.User) error {
	im.users[user.Email] = user
	return nil
}

