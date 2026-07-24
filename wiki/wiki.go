package wiki

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	req.Header.Set("Accept", "text/event-stream")
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
	var data strings.Builder
	for {
		line, err := reader.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")
		switch {
		case trimmed == "":
			dispatchEvent(&data, dataSaver)
			data.Reset()
		case strings.HasPrefix(trimmed, ":"):
			// heartbeat line: ignore.
		default:
			if field, value := parseField(trimmed); field == "data" {
				if data.Len() > 0 {
					// Multiple data lines are joined with newlines per the spec.
					data.WriteByte('\n')
				}
				data.WriteString(value)
			}
		}
		if err != nil {
			return err
		}
	}
}

func parseField(line string) (string, string) {
	field, value, found := strings.Cut(line, ":")
	if !found {
		return line, ""
	}
	return field, strings.TrimPrefix(value, " ")
}

func dispatchEvent(data *strings.Builder, dataSaver func(models.Update)) {
	if data.Len() == 0 {
		return
	}
	update, ok := parseUpdate(data.String())
	if !ok || update.Uri == "" {
		return
	}
	dataSaver(update)
}

func parseUpdate(data string) (models.Update, bool) {
	update := models.Update{}
	if err := json.Unmarshal([]byte(data), &update); err != nil {
		fmt.Println("Error unmarshalling SSE data:", err, "data:", data)
		return models.Update{}, false
	}
	if update.Meta.Uri != "" {
		update.Uri = update.Meta.Uri
	}
	return update, true
}
