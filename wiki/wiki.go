package wiki

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"wiki_updates/configuration"
	"wiki_updates/models"
)

const (
	initialReconnectDelay = 1 * time.Second
	maxReconnectDelay     = 30 * time.Second
)

func ConsumeWikipediaChanges(config configuration.Config, channel chan models.Message) {
	dataSaver := func(update models.Update) {
		channel <- models.Message{
			Type:   "save_data",
			Update: update,
		}
	}
	delay := initialReconnectDelay
	for {
		received, err := streamChanges(config, dataSaver)
		if received {
			delay = initialReconnectDelay
		}
		if err != nil {
			fmt.Printf("Wikipedia stream disconnected, reconnecting in %s: %v\n", delay, err)
		} else {
			fmt.Printf("Wikipedia stream ended, reconnecting in %s\n", delay)
		}
		time.Sleep(delay)
		delay = nextDelay(delay)
	}
}

func nextDelay(current time.Duration) time.Duration {
	next := current * 2
	if next > maxReconnectDelay {
		return maxReconnectDelay
	}
	return next
}

func streamChanges(config configuration.Config, dataSaver func(models.Update)) (bool, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.WikiAPIURL(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", config.UserAgent())
	req.Header.Set("Accept", "application/json")
	rsp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		if cerr := rsp.Body.Close(); cerr != nil {
			fmt.Println("Error closing response body:", cerr)
		}
	}()
	if rsp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	received := false
	countingSaver := func(update models.Update) {
		received = true
		dataSaver(update)
	}
	err = processBody(bufio.NewReader(rsp.Body), countingSaver)
	return received, err
}

func processBody(reader *bufio.Reader, dataSaver func(models.Update)) error {
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			update := processLine(string(line))
			dataSaver(update)
		}
		if err != nil {
			return err
		}
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
