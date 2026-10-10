package threads

import (
	"strconv"
	"time"
)

type Metrics struct {
	Likes   int `json:"likes"`
	Replies int `json:"replies"`
	Reposts int `json:"reposts"`
	Shares  int `json:"shares"`
}

type Profile struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	Bio            string `json:"bio"`
	AvatarURL      string `json:"avatar_url"`
	FollowersCount int    `json:"followers_count"`
	Verified       bool   `json:"verified"`
}

type Post struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Author    Profile   `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Media     []string  `json:"media,omitempty"`
	Metrics   Metrics   `json:"metrics"`
}

func (p *Post) CsvHeaders() []string {
	return []string{
		"id", "url", "author_username", "author_name", "text", "created_at",
		"likes", "replies", "reposts", "shares",
	}
}

func (p *Post) CsvRow() []string {
	return []string{
		p.ID,
		p.URL,
		p.Author.Username,
		p.Author.Name,
		p.Text,
		p.CreatedAt.Format(time.RFC3339),
		strconv.Itoa(p.Metrics.Likes),
		strconv.Itoa(p.Metrics.Replies),
		strconv.Itoa(p.Metrics.Reposts),
		strconv.Itoa(p.Metrics.Shares),
	}
}
