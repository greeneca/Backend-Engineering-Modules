package wiki

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"wiki_updates/configuration"
	"wiki_updates/models"
	"wiki_updates/data"
)


func ConsumeWikipediaChanges(config configuration.Config,) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.WikiAPIURL(), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", config.UserAgent())
	req.Header.Set("Accept", "application/json")
	rsp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := rsp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()
	reader := bufio.NewReader(rsp.Body)
	streamer := data.GetDataStreamer(config)
	dataSaver := func(update models.Update) {
		streamer.Produce(update)
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
	update := models.Update{}
	if err := json.Unmarshal([]byte(line), &update); err != nil {
		fmt.Println("Error unmarshalling JSON:", err, "line:", line)
		return models.Update{}
	}
	if update.Meta.Uri != "" {
		update.Uri = update.Meta.Uri
	}
	return update
}
