package models

type Update struct {
	Uri     string `json:"uri"`
	Bot     bool   `json:"bot"`
	User    string `json:"user"`
	Meta    Meta   `json:"meta"`
}

type Meta struct {
	Uri string `json:"uri"`
}
