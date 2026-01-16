package data

import (
	"testing"
	"wiki_updates/models"
)

func Test_inMemoryData_SaveUpdate(t *testing.T) {
	tests := []struct {
		name string
		update models.Update
	}{
		{
			name: "Test bot user update",
			update: models.Update{
				Uri:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
				Bot:  true,
				User: "BotUser",
			},
		},{
			name: "Test non-bot user update",
			update: models.Update{
				Uri:  "https://en.wikipedia.org/wiki/Special:Diff/0987654321",
				Bot:  false,
				User: "NonBotUser",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := &InMemory{}
			im.Initialize(nil)
			err := im.SaveUpdate(tt.update)
			if err != nil {
				t.Errorf("inMemoryData.SaveUpdate() error = %v, wantErr %v", err, false)
			}
			if im.statistics.Messages != 1 {
				t.Errorf("inMemoryData.Statistics.Messages = %d, want %d", im.statistics.Messages, 1)
			}
			if tt.update.Bot {
				if len(im.bots) != 1 {
					t.Errorf("inMemoryData.Statistics.Bots = %d, want %d", len(im.bots), 1)
}
			} else {
				if len(im.nonBots) != 1 {
					t.Errorf("inMemoryData.Statistics.NonBots = %d, want %d", len(im.nonBots), 1)
				}
			}
		})
	}
}

func Test_inMemoryData_GetStatistics(t *testing.T) {
	tests := []struct {
		name string
		messages int
		urls []string
		botUsers []string
		nonBotUsers []string
		expected models.Statistics
	}{
		{
			name: "Test statistics with no data",
			messages: 0,
			urls: []string{},
			botUsers: []string{},
			nonBotUsers: []string{},
			expected: models.Statistics{
				Messages: 0,
				Urls:     0,
				Bots:     0,
				NonBots:  0,
			},
		},{
			name: "Test statistics with data",
			messages: 3,
			urls: []string{
				"https://en.wikipedia.org/wiki/Special:Diff/1234567890",
				"https://en.wikipedia.org/wiki/Special:Diff/0987654321",
			},
			botUsers: []string{
				"BotUser1",
			},
			nonBotUsers: []string{
				"NonBotUser1",
			},
			expected: models.Statistics{
				Messages: 3,
				Urls:     2,
				Bots:     1,
				NonBots:  1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := &InMemory{}
			im.Initialize(nil)
			im.statistics.Messages = tt.messages
			for _, url := range tt.urls {
				im.urls[url] = true
			}
			for _, botUser := range tt.botUsers {
				im.bots[botUser] = true
			}
			for _, nonBotUser := range tt.nonBotUsers {
				im.nonBots[nonBotUser] = true
			}
			stats, err := im.GetStatistics()
			if err != nil {
				t.Errorf("inMemoryData.GetStatistics() error = %v, wantErr %v", err, false)
			}
			if *stats != tt.expected {
				t.Errorf("inMemoryData.GetStatistics() = %v, want %v", *stats, tt.expected)
			}
		})
	}
}

func Test_inMemoryData_SaveAndGetUser(t *testing.T) {
	user := &models.User{
		Email: "test@test.com",
		Password: "password",
	}
	im := &InMemory{}
	im.Initialize(nil)
	err := im.SaveUser(user)
	if err != nil {
		t.Errorf("inMemoryData.SaveUser() error = %v, wantErr %v", err, false)
	}
	retrievedUser, err := im.GetUserByEmail(user.Email)
	if err != nil {
		t.Errorf("inMemoryData.GetUserByEmail() error = %v, wantErr %v", err, false)
	}
	if *retrievedUser != *user {
		t.Errorf("inMemoryData.GetUserByEmail() = %v, want %v", *retrievedUser, *user)
	}
}
