package configuration

import (
	"encoding/json"
	"os"
)

type Config struct {
	serverPort, wikiAPIURL, userAgent string
}
type internalConfig struct {
	ServerPort, WikiAPIURL, UserAgent string
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

	return config
}


