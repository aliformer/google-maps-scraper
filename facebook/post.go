package facebook

import (
	"strconv"
	"time"
)

type Metrics struct {
	Likes    int `json:"likes"`
	Comments int `json:"comments"`
	Shares   int `json:"shares"`
}

type Profile struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	AvatarURL string `json:"avatar_url"`
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
		"likes", "comments", "shares",
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
		strconv.Itoa(p.Metrics.Comments),
		strconv.Itoa(p.Metrics.Shares),
	}
}
