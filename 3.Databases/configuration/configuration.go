package configuration

import (
	"encoding/json"
	"os"
)

type Config struct {
	serverPort, wikiAPIURL, userAgent, dataStorage, clusterKeyspace string
	clusterHosts []string
	debug bool
}
type internalConfig struct {
	ServerPort, WikiAPIURL, UserAgent, DataStorage, ClusterKeyspace string
	ClusterHosts []string
	Debug bool
}

func (c *Config) ServerPort() string {
	return c.serverPort
}
func (c *Config) WikiAPIURL() string {
	return c.wikiAPIURL
}
func (c *Config) UserAgent() string {
	return c.userAgent
}
func (c *Config) DataStorage() string {
	return c.dataStorage
}
func (c *Config) ClusterHosts() []string {
	return c.clusterHosts
}
func (c *Config) ClusterKeyspace() string {
	return c.clusterKeyspace
}
func (c *Config) Debug() bool {
	return c.debug
}

func GetConfig() Config {
	file, err := os.Open("wiki_updates.conf.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	internalConfig := internalConfig{}
	err = decoder.Decode(&internalConfig)
	if err != nil {
		panic(err)
	}

	// Set default values
	config := Config{
		serverPort: "7000",
		wikiAPIURL: "https://stream.wikimedia.org/v2/stream/recentchange",
		userAgent:  "WikiUpdatesBot/0.0 (charles.greene@redspace.com) go/1.24.5",
		dataStorage:    "memory", // or "cassandra"
		clusterHosts: []string{"database"},
		clusterKeyspace: "wiki_updates",
		debug: false,
	}
	if internalConfig.ServerPort != "" {
		config.serverPort = internalConfig.ServerPort
	}
	if internalConfig.WikiAPIURL != "" {
		config.wikiAPIURL = internalConfig.WikiAPIURL
	}
	if internalConfig.UserAgent != "" {
		config.userAgent = internalConfig.UserAgent
	}
	if internalConfig.DataStorage != "" {
		if internalConfig.DataStorage == "cassandra" || internalConfig.DataStorage == "memory" {
			config.dataStorage = internalConfig.DataStorage
		}
	}
	if len(internalConfig.ClusterHosts) > 0 {
		config.clusterHosts = internalConfig.ClusterHosts
	}
	if internalConfig.ClusterKeyspace != "" {
		config.clusterKeyspace = internalConfig.ClusterKeyspace
	}
	if internalConfig.Debug {
		config.debug = internalConfig.Debug
	}

	return config
}


