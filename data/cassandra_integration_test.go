package data

import (
	"context"
	"log"
	//"path/filepath"
	"testing"
	mock_configuration "wiki_updates/configuration/mock"
	"wiki_updates/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	//"github.com/testcontainers/testcontainers-go/modules/cassandra"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB() testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "cassandra:latest",
		ExposedPorts: []string{"9042/tcp"},
		Env: map[string]string{
			"CASSANDRA_BROADCAST_ADDRESS": "localhost",
		},
		WaitingFor:   wait.ForAll(wait.ForLog("state jump to NORMAL"), wait.ForListeningPort("9042/tcp")),
		Name:         "cassy_integration_test",
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start scylla: %s", err)
	}
	exec, i, err := container.Exec(context.Background(), []string{"cqlsh", "-e", "init.sh"}, tcexec.Multiplexed())
	if err != nil {
		log.Fatal(exec, i, err)
	}
	return container
}

func TestIntegration_Cassendra(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctrl := gomock.NewController(t)
	//ctx := context.Background()
	//cassandraContainer := setupTestDB()
	//cassandraContainer, err := cassandra.Run(
	//	ctx,
	//	"cassandra:latest",
	//	cassandra.WithInitScripts(filepath.Join("scripts", "init.sh")),
	//	testcontainers.WithEnv(map[string]string{"CASSANDRA_BROADCAST_ADDRESS": "localhost"}),
	//)
	//assert.NoError(t, err)
	//defer testcontainers.TerminateContainer(cassandraContainer)

	//endpoint, err := cassandraContainer.Endpoint(ctx, "")
	//if err != nil {
//		t.Fatal("Failed to get Cassandra endpoint:", err)
//	}
//	println("Cassandra endpoint: '", endpoint, "'")
//	assert.NoError(t, err)
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
