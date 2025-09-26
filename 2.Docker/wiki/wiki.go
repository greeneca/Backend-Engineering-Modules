package wiki

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/statistics"
)


func ConsumeWikipediaChanges(stats *statistics.Statistics, config configuration.Config) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.WikiAPIURL(), nil)
	req.Header.Set("User-Agent", config.UserAgent())
	req.Header.Set("Accept", "application/json")
	rsp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	reader := bufio.NewReader(rsp.Body)
	processBody(reader, stats)
}

func processBody(reader *bufio.Reader, stats *statistics.Statistics) {
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		if len(line) ==0 {
			continue
		}
		processLine(string(line), stats)
	}

}

func processLine(line string, stats *statistics.Statistics) {
	jsonData := make(map[string]any)
	if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
		fmt.Println("Error unmarshalling JSON:", err, "line:", line)
		return
	}
	stats.Lock()
	stats.Messages++
	stats.Unlock()
	if meta, ok := jsonData["meta"].(map[string]any); ok {
		if uri, ok := meta["uri"].(string); ok {
			stats.Lock()
			stats.Urls[uri] = true
			stats.Unlock()
		}
	}
	if bot, ok := jsonData["bot"].(bool); ok {
		if bot {
			if user, ok := jsonData["user"].(string); ok {
				stats.Lock()
				stats.Bots[user] = true
				stats.Unlock()
			}
		} else {
			if user, ok := jsonData["user"].(string); ok {
				stats.Lock()
				stats.NonBots[user] = true
				stats.Unlock()
			}
		}
	}
}
