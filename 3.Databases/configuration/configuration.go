package configuration

import (
	"encoding/json"
	"fmt"
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

func defaultConfig() Config {
	config := Config{
		serverPort: "7000",
		wikiAPIURL: "https://stream.wikimedia.org/v2/stream/recentchange",
		userAgent:  "WikiUpdatesBot/0.0 (charles.greene@redspace.com) go/1.24.5",
		dataStorage:    "memory", // or "cassandra"
		clusterHosts: []string{"database"},
		clusterKeyspace: "wiki_updates",
		debug: false,
	}
	return config
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
	internalConfig := loadConfigFromFile("wiki_updates.conf.json")
	config := defaultConfig()
	updateConfigWithInternalConfig(&config, internalConfig)
	return config
}

func loadConfigFromFile(filename string) internalConfig {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening config file, using default configuration. (%v)\n", err)
		return internalConfig{}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	internalConf := internalConfig{}
	err = decoder.Decode(&internalConf)
	if err != nil {
		fmt.Printf("Error decoding config file, using default configuration. (%v)\n", err)
		return internalConfig{}
	}
	return internalConf
}

func updateConfigWithInternalConfig(config *Config, internalConfig internalConfig) {
	// Set default values
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
}


