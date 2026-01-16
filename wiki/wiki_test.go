package wiki

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"wiki_updates/models"
)

func Test_processLine(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		line string
		Url string
		Bot bool
		User string
	}{
		{
			name: "Test empty data line",
			line: "",
			Url: "",
			Bot: false,
			User: "",
		},{
			name: "Test data line with message",
			line: "{}",
			Url: "",
			Bot: false,
			User: "",
		},{
			name: "Test data line with bot",
			line: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/1234567890"},"bot":true,"user":"BotUser"}`,
			Url: "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
			Bot: true,
			User: "BotUser",
		},{
			name: "Test data line with non-bot",
			line: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/0987654321"},"bot":false,"user":"NonBotUser"}`,
			Url: "https://en.wikipedia.org/wiki/Special:Diff/0987654321",
			Bot: false,
			User: "NonBotUser",
		},{
			name: "Test data line with uri",
			line: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/1122334455"}}`,
			Url: "https://en.wikipedia.org/wiki/Special:Diff/1122334455",
			Bot: false,
			User: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := processLine(tt.line)
			if got, want := update.Uri, tt.Url; got != want {
				t.Errorf("processBody() Url = %s, want %s", got, want)
			}
			if got, want := update.Bot, tt.Bot; got != want {
				t.Errorf("processBody() Bot = %t, want %t", got, want)
			}
			if got, want := update.User, tt.User; got != want {
				t.Errorf("processBody() User = %s, want %s", got, want)
			}
		})
	}
}

func Test_processBody(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		reader *bufio.Reader
		foundUpdates int
		updates int
	}{
		{
			name: "Test processBody with valid data",
			reader: bufio.NewReader(strings.NewReader("{}")),
			updates: 1,
		},{
			name: "Test processBody with multiple lines",
			reader: bufio.NewReader(strings.NewReader("{}\n{}\n")),
			updates: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != io.EOF {
					t.Errorf("processBody() panicked with error: %v", r)
				}
			}()
			dataSaver := func(update models.Update) {
				tt.foundUpdates++
			}
			processBody(tt.reader, dataSaver)
			if got, want := tt.foundUpdates, tt.updates; got != want {
				t.Errorf("processBody() Messages = %d, want %d", got, want)
			}
		})
	}
}

