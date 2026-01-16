package models

type Message struct {
	Type	string
	Update	Update
	Statistics	*Statistics
	User	User
	Error	error
}
