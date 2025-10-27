package models

type Statistics struct {
	Messages int `json:"messages"`
	Urls int `json:"urls"`
	Bots int `json:"bots"`
	NonBots int `json:"non_bots"`
}
