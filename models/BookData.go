package models

import "time"

type BookData struct {
	Title         string
	Subtitle      string
	Description   string
	Publisher     string
	Thumbnail     string
	PublishedDate time.Time
	SelfLink      string
	Categories    string
	Authors       string
	ISBN10        string
	ISBN13        string
	Source        string
}
