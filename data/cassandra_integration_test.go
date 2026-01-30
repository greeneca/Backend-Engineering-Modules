package data

import (
	"testing"
	mock_configuration "wiki_updates/configuration/mock"
	"wiki_updates/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Cassendra(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctrl := gomock.NewController(t)
	config := mock_configuration.NewMockConfig(ctrl)
	config.EXPECT().ClusterHosts().Return([]string{"database"}).AnyTimes()
	config.EXPECT().Debug().Return(false).AnyTimes()
	config.EXPECT().ClusterKeyspace().Return("wiki_updates").AnyTimes()
	db := Cassandra{}
	db.Initialize(config)

	// Test SaveUpdate
	update := models.Update{
		Uri:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
		Bot:  true,
		User: "IntegrationTestBot",
	}
	err := db.SaveUpdate(update)
	assert.NoError(t, err)

	// Test GetStatistics
	stats, err := db.GetStatistics()
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Messages)
	assert.Equal(t, 1, stats.Urls)
	assert.Equal(t, 1, stats.Bots)
	assert.Equal(t, 0, stats.NonBots)

	// Test SaveUser and GetUserByEmail
	test_user := &models.User{
		Email: "test@user.com",
		PasswordHash: "hashed_password",
	}
	err = db.SaveUser(test_user)
	assert.NoError(t, err)
	retrievedUser, err := db.GetUserByEmail(test_user.Email)
	assert.NoError(t, err)
	assert.Equal(t, test_user.Email, retrievedUser.Email)
	assert.Equal(t, test_user.PasswordHash, retrievedUser.PasswordHash)
}
