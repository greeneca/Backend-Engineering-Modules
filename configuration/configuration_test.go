package configuration

import (
	"reflect"
	"testing"
)

func Test_updateConfigWithInternalConfig(t *testing.T) {
	tests := []struct {
		name string
		want Configuration
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
				JWTSecret: "test_secret",
				Debug: true,
			},
			want: Configuration{
				serverPort: "8000",
				wikiAPIURL: "https://test.url.com/stream",
				userAgent: "TestAgent/1.0",
				dataStorage: "cassandra",
				clusterHosts: []string{"host1", "host2"},
				clusterKeyspace: "test_keyspace",
				jwtSecret: "test_secret",
				debug: true,
			},
		},
		{
			name: "Test Partial Data",
			internalConfig: internalConfig{
				ServerPort: "8000",
				Debug: true,
			},
			want: Configuration{
				serverPort: "8000",
				wikiAPIURL: "https://stream.wikimedia.org/v2/stream/recentchange",
				userAgent:  "WikiUpdatesBot/0.0 (charles.greene@redspace.com) go/1.24.5",
				dataStorage:    "memory",
				clusterHosts: []string{"database"},
				clusterKeyspace: "wiki_updates",
				jwtSecret: "",
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

func Test_getJWTSecret(t *testing.T) {
	t.Run("env var takes precedence over config file", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "env_secret")
		config := Configuration{jwtSecret: "file_secret", debug: false}
		getJWTSecret(&config)
		if config.jwtSecret != "env_secret" {
			t.Errorf("jwtSecret = %q, want %q", config.jwtSecret, "env_secret")
		}
	})

	t.Run("falls back to config file secret when env unset", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		config := Configuration{jwtSecret: "file_secret", debug: false}
		getJWTSecret(&config)
		if config.jwtSecret != "file_secret" {
			t.Errorf("jwtSecret = %q, want %q", config.jwtSecret, "file_secret")
		}
	})

	t.Run("uses dev secret in debug mode when nothing else set", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		config := Configuration{jwtSecret: "", debug: true}
		getJWTSecret(&config)
		if config.jwtSecret != devJWTSecret {
			t.Errorf("jwtSecret = %q, want %q", config.jwtSecret, devJWTSecret)
		}
	})

	t.Run("panics when no secret and not debug", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		config := Configuration{jwtSecret: "", debug: false}
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("getJWTSecret() did not panic with missing secret outside debug mode")
			}
		}()
		getJWTSecret(&config)
	})
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
				JWTSecret: "custom_secret",
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
