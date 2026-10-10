package tiktok

import (
	"strconv"
	"time"
)

type Metrics struct {
	Likes    int `json:"likes"`
	Comments int `json:"comments"`
	Shares   int `json:"shares"`
	Views    int `json:"views"`
}

type Profile struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Video struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Author      Profile   `json:"author"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Metrics     Metrics   `json:"metrics"`
}

func (v *Video) CsvHeaders() []string {
	return []string{
		"id", "url", "author_username", "author_name", "description", "created_at",
		"likes", "comments", "shares", "views",
	}
}

func (v *Video) CsvRow() []string {
	return []string{
		v.ID,
		v.URL,
		v.Author.Username,
		v.Author.Name,
		v.Description,
		v.CreatedAt.Format(time.RFC3339),
		strconv.Itoa(v.Metrics.Likes),
		strconv.Itoa(v.Metrics.Comments),
		strconv.Itoa(v.Metrics.Shares),
		strconv.Itoa(v.Metrics.Views),
	}
}
