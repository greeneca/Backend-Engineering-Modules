package configuration

import (
	"reflect"
	"testing"
)

func Test_updateConfigWithInternalConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
		internalConfig internalConfig
	}{
		{
			name: "Test Basic Data",
			internalConfig: internalConfig{
				ServerPort: "8000",
				WikiAPIURL: "https://test.url.com/stream",
				UserAgent: "TestAgent/1.0",
				DataStorage: "cassandra",
				ClusterHosts: []string{"host1", "host2"},
				ClusterKeyspace: "test_keyspace",
				Debug: true,
			},
			want: Config{
				serverPort: "8000",
				wikiAPIURL: "https://test.url.com/stream",
				userAgent: "TestAgent/1.0",
				dataStorage: "cassandra",
				clusterHosts: []string{"host1", "host2"},
				clusterKeyspace: "test_keyspace",
				debug: true,
			},
		},
		{
			name: "Test Partial Data",
			internalConfig: internalConfig{
				ServerPort: "8000",
				Debug: true,
			},
			want: Config{
				serverPort: "8000",
				wikiAPIURL: "https://stream.wikimedia.org/v2/stream/recentchange",
				userAgent:  "WikiUpdatesBot/0.0 (charles.greene@redspace.com) go/1.24.5",
				dataStorage:    "memory",
				clusterHosts: []string{"database"},
				clusterKeyspace: "wiki_updates",
				debug: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultConfig()
			updateConfigWithInternalConfig(&got, tt.internalConfig)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateConfigWithInternalConfig = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultConfig(t *testing.T) {
	tests := []struct {
		name string
		want internalConfig
		fileName string
	}{
		{
			name: "Test no file",
			want: internalConfig{},
			fileName: "non_existent_file.json",
		},
		{
			name: "Test empty file",
			want: internalConfig{},
			fileName: "test_files/test_empty_config.json",
		},
		{
			name: "Test full config file",
			want: internalConfig{
				ServerPort: "9000",
				WikiAPIURL: "https://custom.url/stream",
				UserAgent: "CustomAgent/2.0",
				DataStorage: "cassandra",
				ClusterHosts: []string{"custom_host1", "custom_host2"},
				ClusterKeyspace: "custom_keyspace",
				Debug: true,
			},
			fileName: "test_files/test_full_config.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loadConfigFromFile(tt.fileName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadConfigFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
