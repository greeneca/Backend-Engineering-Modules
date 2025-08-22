package wiki

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wiki_updates/statstics"
)


func ConsumeWikipediaChanges(stats *statstics.Statstics) {
	rsp, err := http.Get("https://stream.wikimedia.org/v2/stream/recentchange")
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	reader := bufio.NewReader(rsp.Body)
	processBody(reader, stats)
}

func processBody(reader *bufio.Reader, stats *statstics.Statstics) {
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		processLine(string(line), stats)
	}

}

func processLine(line string, stats *statstics.Statstics) {
	if line == "event: message\n" {
		stats.Lock()
		stats.Messages++
		stats.Unlock()
	} else if strings.HasPrefix(line, "data: ") {
		data := strings.TrimPrefix(line, "data: ")
		jsonData := make(map[string]any)
		if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}
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
}
