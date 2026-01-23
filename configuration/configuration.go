package configuration

//go:generate mockgen -source=configuration.go -destination=mock/configuration.go

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config interface {
	ServerPort() string
	WikiAPIURL() string
	UserAgent() string
	DataStorage() string
	ClusterHosts() []string
	ClusterKeyspace() string
	JWTSecret() string
	Debug() bool
}

type Configuration struct {
	serverPort, wikiAPIURL, userAgent, dataStorage, clusterKeyspace, jwtSecret string
	clusterHosts []string
	debug bool
}
type internalConfig struct {
	ServerPort, WikiAPIURL, UserAgent, DataStorage, ClusterKeyspace, JWTSecret string
	ClusterHosts []string
	Debug bool
}

func defaultConfig() Configuration {
	config := Configuration{
		serverPort: "7000",
		wikiAPIURL: "https://stream.wikimedia.org/v2/stream/recentchange",
		userAgent:  "WikiUpdatesBot/0.0 (charles.greene@redspace.com) go/1.24.5",
		dataStorage:    "memory", // or "cassandra"
		clusterHosts: []string{"database"},
		clusterKeyspace: "wiki_updates",
		jwtSecret: "supersecretkey",
		debug: false,
	}
	return config
}


func (c *Configuration) ServerPort() string {
	return c.serverPort
}
func (c *Configuration) WikiAPIURL() string {
	return c.wikiAPIURL
}
func (c *Configuration) UserAgent() string {
	return c.userAgent
}
func (c *Configuration) DataStorage() string {
	return c.dataStorage
}
func (c *Configuration) ClusterHosts() []string {
	return c.clusterHosts
}
func (c *Configuration) ClusterKeyspace() string {
	return c.clusterKeyspace
}
func (c *Configuration) JWTSecret() string {
	return c.jwtSecret
}
func (c *Configuration) Debug() bool {
	return c.debug
}

func GetConfig() Configuration {
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
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing config file. (%v)\n", err)
		}
	}()
	decoder := json.NewDecoder(file)
	internalConf := internalConfig{}
	err = decoder.Decode(&internalConf)
	if err != nil {
		fmt.Printf("Error decoding config file, using default configuration. (%v)\n", err)
		return internalConfig{}
	}
	return internalConf
}

func updateConfigWithInternalConfig(config *Configuration, internalConfig internalConfig) {
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
	if internalConfig.JWTSecret != "" {
		config.jwtSecret = internalConfig.JWTSecret
	}
	if internalConfig.Debug {
		config.debug = internalConfig.Debug
	}
}


