package wiki

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/models"
)


func ConsumeWikipediaChanges(config configuration.Config, channel *chan models.Message) {
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
	dataSaver := func(update models.Update) {
		*channel <- models.Message{
			Type:   "save_data",
			Update: update,
		}
	}
	processBody(reader, dataSaver)
}

func processBody(reader *bufio.Reader, dataSaver func(models.Update)) {
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		if len(line) ==0 {
			continue
		}
		update := processLine(string(line))
		dataSaver(update)
	}

}

func processLine(line string) models.Update {
	jsonData := make(map[string]any)
	if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
		fmt.Println("Error unmarshalling JSON:", err, "line:", line)
		return models.Update{}
	}
	update := models.Update{}

	if meta, ok := jsonData["meta"].(map[string]any); ok {
		if uri, ok := meta["uri"].(string); ok {
			update.Uri = uri
		}
		if bot, ok := jsonData["bot"].(bool); ok {
			update.Bot = bot
		}
		if user, ok := jsonData["user"].(string); ok {
			update.User = user
		}
	}
	return update
}
