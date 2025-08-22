package statstics

import "sync"

type Statstics struct {
	sync.RWMutex
	Messages int `json:"messages"`
	Urls map[string]bool `json:"urls"`
	Bots map[string]bool `json:"bots"`
	NonBots map[string]bool `json:"non_bots"`
}
