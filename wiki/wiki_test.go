package wiki

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"time"
	"wiki_updates/models"
)

func Test_parseUpdate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		data string
		ok   bool
		Url  string
		Bot  bool
		User string
	}{
		{
			name: "Test empty data",
			data: "",
			ok:   false,
		},{
			name: "Test malformed data",
			data: "not json",
			ok:   false,
		},{
			name: "Test empty object",
			data: "{}",
			ok:   true,
			Url:  "",
		},{
			name: "Test data with bot",
			data: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/1234567890"},"bot":true,"user":"BotUser"}`,
			ok:   true,
			Url:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
			Bot:  true,
			User: "BotUser",
		},{
			name: "Test data with non-bot",
			data: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/0987654321"},"bot":false,"user":"NonBotUser"}`,
			ok:   true,
			Url:  "https://en.wikipedia.org/wiki/Special:Diff/0987654321",
			Bot:  false,
			User: "NonBotUser",
		},{
			name: "Test data with uri",
			data: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/1122334455"}}`,
			ok:   true,
			Url:  "https://en.wikipedia.org/wiki/Special:Diff/1122334455",
			Bot:  false,
			User: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update, ok := parseUpdate(tt.data)
			if ok != tt.ok {
				t.Fatalf("parseUpdate() ok = %t, want %t", ok, tt.ok)
			}
			if !ok {
				return
			}
			if got, want := update.Uri, tt.Url; got != want {
				t.Errorf("parseUpdate() Url = %s, want %s", got, want)
			}
			if got, want := update.Bot, tt.Bot; got != want {
				t.Errorf("parseUpdate() Bot = %t, want %t", got, want)
			}
			if got, want := update.User, tt.User; got != want {
				t.Errorf("parseUpdate() User = %s, want %s", got, want)
			}
		})
	}
}

func Test_processBody(t *testing.T) {
	valid := `{"meta":{"uri":"https://en.wikipedia.org/wiki/A"},"bot":true,"user":"Bot"}`
	other := `{"meta":{"uri":"https://en.wikipedia.org/wiki/B"}}`
	tests := []struct {
		name    string // description of this test case
		body    string
		updates int
	}{
		{
			name:    "single event",
			body:    "data: " + valid + "\n\n",
			updates: 1,
		},{
			name:    "multiple events",
			body:    "data: " + valid + "\n\ndata: " + other + "\n\n",
			updates: 2,
		},{
			name:    "comments and metadata fields ignored",
			body:    ":ok\n\nevent: message\nid: 42\ndata: " + valid + "\n\n",
			updates: 1,
		},{
			name:    "malformed data not saved",
			body:    "data: not json\n\n",
			updates: 0,
		},{
			name:    "empty object not saved",
			body:    "data: {}\n\n",
			updates: 0,
		},{
			name:    "unterminated event discarded",
			body:    "data: " + valid,
			updates: 0,
		},{
			name:    "multi-line data concatenated",
			body:    "data: {\"meta\":{\"uri\":\"https://x/D\"},\ndata: \"bot\":true}\n\n",
			updates: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUpdates := 0
			dataSaver := func(update models.Update) {
				foundUpdates++
			}
			reader := bufio.NewReader(strings.NewReader(tt.body))
			if err := processBody(reader, dataSaver); err != io.EOF {
				t.Errorf("processBody() error = %v, want %v", err, io.EOF)
			}
			if got, want := foundUpdates, tt.updates; got != want {
				t.Errorf("processBody() updates = %d, want %d", got, want)
			}
		})
	}
}

func Test_nextDelay(t *testing.T) {
	tests := []struct {
		name    string
		current time.Duration
		want    time.Duration
	}{
		{
			name:    "doubles below cap",
			current: 1 * time.Second,
			want:    2 * time.Second,
		},{
			name:    "doubles up to cap",
			current: maxReconnectDelay / 2,
			want:    maxReconnectDelay,
		},{
			name:    "stays at cap",
			current: maxReconnectDelay,
			want:    maxReconnectDelay,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextDelay(tt.current); got != tt.want {
				t.Errorf("nextDelay(%s) = %s, want %s", tt.current, got, tt.want)
			}
		})
	}
}

