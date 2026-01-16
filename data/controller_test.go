package data

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	mock_configuration "wiki_updates/configuration/mock"
	mock_data "wiki_updates/data/mock"
	"wiki_updates/models"

	"github.com/golang/mock/gomock"
)

func Test_Controller_getDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name   string
		setExpectations func(m *mock_configuration.MockConfig)
		want   DataSource
	}{
		{
			name: "Test memory data source",
			setExpectations: func(m *mock_configuration.MockConfig) {
				m.EXPECT().DataStorage().Return("memory")
			},
			want: &InMemory{},
		},
		{
			name: "Test cassandra data source",
			setExpectations: func(m *mock_configuration.MockConfig) {
				m.EXPECT().DataStorage().Return("cassandra")
			},
			want: &Cassandra{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := mock_configuration.NewMockConfig(ctrl)
			tt.setExpectations(mockConfig)
			got := getDataSource(mockConfig)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("getDataSource() = %T, want %T", got, tt.want)
			}
		})
	}
}

func Test_Controller_monitorChannels(t *testing.T) {
	server_tests := []struct {
		name     string
		input    models.Message
		expected models.Message
	}{
		{
			name: "Test get_stats message",
			input: models.Message{
				Type: "get_stats",
			},
			expected: models.Message{
				Type:       "stats_response",
				Statistics: &models.Statistics{
					Messages: 1,
					Urls:     2,
					Bots:     3,
					NonBots:  4,
				},
			},
		},
	}
	for _, tt := range server_tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			wikiChan := make(chan models.Message)
			serverChan := make(chan models.Message)
			dataSource := mock_data.NewMockDataSource(ctrl)
			dataSource.EXPECT().GetStatistics().Return(tt.expected.Statistics, nil)
			go monitorChannels(&wikiChan, &serverChan, dataSource)
			serverChan <- tt.input
			got := <-serverChan
			if got.Type != tt.expected.Type {
				fmt.Printf("got: %+v\n", got)
				fmt.Printf("expected: %+v\n", tt.expected)
				t.Errorf("monitorChannels() Type = %s, want %s", got.Type, tt.expected.Type)
			}
			if got.Statistics.Messages != tt.expected.Statistics.Messages ||
				got.Statistics.Urls != tt.expected.Statistics.Urls ||
				got.Statistics.Bots != tt.expected.Statistics.Bots ||
				got.Statistics.NonBots != tt.expected.Statistics.NonBots {
				t.Errorf("monitorChannels() Statistics = %v, want %v", got.Statistics, tt.expected.Statistics)
			}
		})
	}
	wiki_tests := []struct {
		name  string
		inputs []models.Message
		expectedUpdates int
	}{
		{
			name: "Test single save_data messages",
			inputs: []models.Message{
				{Type: "save_data", Update: models.Update{}},
			},
			expectedUpdates: 1,
		},
		{
			name: "Test multiple save_data messages",
			inputs: []models.Message{
				{Type: "save_data", Update: models.Update{}},
				{Type: "save_data", Update: models.Update{}},
			},
			expectedUpdates: 2,
		},
	}
	for _, tt := range wiki_tests {
		t.Run(tt.name, func(t *testing.T) {
			wikiChan := make(chan models.Message)
			serverChan := make(chan models.Message)
			waitGroup := sync.WaitGroup{}
			waitGroup.Add(tt.expectedUpdates)
			dataSource := mock_data.NewMockDataSource(gomock.NewController(t))
			dataSource.EXPECT().SaveUpdate(gomock.Any()).Times(tt.expectedUpdates).DoAndReturn(func(_ models.Update) error {
				waitGroup.Done()
				return nil
			})
			go monitorChannels(&wikiChan, &serverChan, dataSource)
			for _, input := range tt.inputs {
				wikiChan <- input
			}
			waitGroup.Wait()
		})
	}
}
