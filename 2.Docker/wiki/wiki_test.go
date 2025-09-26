package wiki

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"wiki_updates/statistics"
)

func Test_processLine(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		line string
		stats  statistics.Statistics
		messages int
		Urls int
		Bots int
		NonBots int
	}{
		{
			name: "Test empty data line",
			line: "",
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 0,
			Urls: 0,
			Bots: 0,
			NonBots: 0,
		},{
			name: "Test data line with message",
			line: "{}",
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 1,
			Urls: 0,
			Bots: 0,
			NonBots: 0,
		},{
			name: "Test data line with bot",
			line: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/1234567890"},"bot":true,"user":"BotUser"}`,
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 1,
			Urls: 1,
			Bots: 1,
			NonBots: 0,
		},{
			name: "Test data line with non-bot",
			line: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/0987654321"},"bot":false,"user":"NonBotUser"}`,
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 1,
			Urls: 1,
			Bots: 0,
			NonBots: 1,
		},{
			name: "Test data line with uri",
			line: `{"meta":{"uri":"https://en.wikipedia.org/wiki/Special:Diff/1122334455"}}`,
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 1,
			Urls: 1,
			Bots: 0,
			NonBots: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processLine(tt.line, &tt.stats)
			if got, want := tt.stats.Messages, tt.messages; got != want {
				t.Errorf("processBody() Messages = %d, want %d", got, want)
			}
			if got, want := len(tt.stats.Urls), tt.Urls; got != want {
				t.Errorf("processBody() Urls = %d, want %d", got, want)
			}
			if got, want := len(tt.stats.Bots), tt.Bots; got != want {
				t.Errorf("processBody() Bots = %d, want %d", got, want)
			}
			if got, want := len(tt.stats.NonBots), tt.NonBots; got != want {
				t.Errorf("processBody() NonBots = %d, want %d", got, want)
			}
		})
	}
}

func Test_processBody(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		reader *bufio.Reader
		stats  statistics.Statistics
		messages int
	}{
		{
			name: "Test processBody with valid data",
			reader: bufio.NewReader(strings.NewReader("{}")),
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 1,
		},{
			name: "Test processBody with multiple lines",
			reader: bufio.NewReader(strings.NewReader("{}\n{}\n")),
			stats: statistics.Statistics{Messages: 0,Urls: make(map[string]bool), Bots: make(map[string]bool), NonBots: make(map[string]bool),},
			messages: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != io.EOF {
					t.Errorf("processBody() panicked with error: %v", r)
				}
			}()
			processBody(tt.reader, &tt.stats)
			if got, want := tt.stats.Messages, tt.messages; got != want {
				t.Errorf("processBody() Messages = %d, want %d", got, want)
			}
		})
	}
}

