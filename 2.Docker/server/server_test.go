package server

import(
	"github.com/gin-gonic/gin"
	"testing"
	"wiki_updates/statstics"
)

func Test_generateStatsJson(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		stats *statstics.Statstics
		want  gin.H
	}{
		{
			name: "Test empty stats",
			stats: &statstics.Statstics{
				Messages: 0,
				Urls:     make(map[string]bool),
				Bots:     make(map[string]bool),
				NonBots:  make(map[string]bool),
			},
			want: gin.H{
				"messages": 0,
				"urls":     0,
				"users":    0,
				"bots":     0,
				"non_bots": 0,
			},
		},{
			name: "Test stats with messages and urls",
			stats: &statstics.Statstics{
				Messages: 5,
				Urls: map[string]bool{
					"https://en.wikipedia.org/wiki/Special:Diff/1234567890": true,
					"https://en.wikipedia.org/wiki/Special:Diff/0987654321": true,
				},
				Bots: map[string]bool{
					"BotUser": true,
				},
				NonBots: map[string]bool{
					"NonBotUser": true,
				},
			},
			want: gin.H{
				"messages": 5,
				"urls":     2,
				"users":    2,
				"bots":     1,
				"non_bots": 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := generateStatsJson(tt.stats), tt.want; got["messages"] != want["messages"] || got["urls"] != want["urls"] || got["users"] != want["users"] || got["bots"] != want["bots"] || got["non_bots"] != want["non_bots"] {
				t.Errorf("generateStatsJson() = %v, want %v", got, want)
			}
		})
	}
}

