package models

import "time"

type Feed struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"` // New field
	LastUpdated time.Time `json:"last_updated"`
}

type Article struct {
	ID              int64     `json:"id"`
	FeedID          int64     `json:"feed_id"`
	Title           string    `json:"title"`
	URL             string    `json:"url"`
	ImageURL        string    `json:"image_url"`
	PublishedAt     time.Time `json:"published_at"`
	IsRead          bool      `json:"is_read"`
	IsFavorite      bool      `json:"is_favorite"`
	FeedTitle       string    `json:"feed_title,omitempty"` // Joined field
	TranslatedTitle string    `json:"translated_title"`
}
