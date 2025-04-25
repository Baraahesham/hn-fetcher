package models

import (
	"time"
)

type Story struct {
	HnID   int64  `json:"id"`     // Hacker News ID, also primary key in DB
	Title  string `json:"title"`  // Story title
	Author string `json:"author"` // Hacker News "by" field
	URL    string `json:"url"`    // Optional: external link
	Time   int64  `json:"time"`   // Parsed Unix timestamp from HN
}

type StoryDbModel struct {
	ID     int64 `gorm:"primaryKey"`
	Title  string
	Author string
	URL    string
	Time   time.Time
	HnID   int64 `gorm:"uniqueIndex"` // Unique index for Hacker News ID
}

func (StoryDbModel) TableName() string {
	return "stories"
}

type StoryEvent struct {
	HnID  int64  `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}
type BrandStats struct {
	Brand    string `json:"brand"`
	Mentions int    `json:"mentions"`
}
