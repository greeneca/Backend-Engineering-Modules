package data

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
	mock_configuration "wiki_updates/configuration/mock"
	"wiki_updates/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_Redpanda runs against a real broker. Set REDPANDA_BROKERS
// (comma separated) to enable it.
func TestIntegration_Redpanda(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	brokers := os.Getenv("REDPANDA_BROKERS")
	if brokers == "" {
		t.Skip("REDPANDA_BROKERS not set; skipping Redpanda integration test")
	}

	// Unique topics keep each run isolated from earlier runs.
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	ctrl := gomock.NewController(t)
	config := mock_configuration.NewMockConfig(ctrl)
	config.EXPECT().RedpandaBrokers().Return(strings.Split(brokers, ",")).AnyTimes()
	config.EXPECT().RedpandaTopic().Return("it.updates." + suffix).AnyTimes()
	config.EXPECT().RedpandaUsersTopic().Return("it.users." + suffix).AnyTimes()
	config.EXPECT().RedpandaSnapshotTopic().Return("it.snapshot." + suffix).AnyTimes()

	db := &Redpanda{}
	db.Initialize(config)
	defer db.streamer.Close()

	// SaveUpdate goes through the broker and comes back via the consumer.
	update := models.Update{
		Uri:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
		Bot:  true,
		User: "IntegrationTestBot",
	}
	assert.NoError(t, db.SaveUpdate(update))

	assert.Eventually(t, func() bool {
		stats, err := db.GetStatistics()
		return err == nil && stats.Messages == 1 && stats.Urls == 1 && stats.Bots == 1 && stats.NonBots == 0
	}, 30*time.Second, 200*time.Millisecond, "update was not materialized from the updates topic")

	// SaveUser and GetUserByEmail
	user := &models.User{Email: "test@user.com", PasswordHash: "hashed_password"}
	assert.NoError(t, db.SaveUser(user))
	got, err := db.GetUserByEmail(user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, got.Email)
	assert.Equal(t, user.PasswordHash, got.PasswordHash)

	// Duplicate users are rejected.
	assert.Error(t, db.SaveUser(user))
}
