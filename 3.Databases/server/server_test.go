package server

import (
	"testing"
	"wiki_updates/models"

	"github.com/gin-gonic/gin"
)

func Test_generateStatsJson(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		statsFetcher func() *models.Statistics
		want  gin.H
	}{
		{
			name: "Test empty stats",
			statsFetcher: func() *models.Statistics {
				return &models.Statistics{
					Messages: 0,
					Urls:     0,
					Bots:     0,
					NonBots:  0,
				}
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
			statsFetcher: func() *models.Statistics {
				return &models.Statistics{
					Messages: 5,
					Urls:     2,
					Bots:     1,
					NonBots:  1,
				}
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
			if got, want := generateStatsJson(tt.statsFetcher), tt.want; got["messages"] != want["messages"] || got["urls"] != want["urls"] || got["users"] != want["users"] || got["bots"] != want["bots"] || got["non_bots"] != want["non_bots"] {
				t.Errorf("generateStatsJson() = %v, want %v", got, want)
			}
		})
	}
}

