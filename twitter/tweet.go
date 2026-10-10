package twitter

import (
	"strconv"
	"time"
)

type Metrics struct {
	Replies   int `json:"replies"`
	Retweets  int `json:"retweets"`
	Likes     int `json:"likes"`
	Bookmarks int `json:"bookmarks"`
	Views     int `json:"views"`
}

type Profile struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Name           string    `json:"name"`
	Bio            string    `json:"bio"`
	URL            string    `json:"url"`
	AvatarURL      string    `json:"avatar_url"`
	BannerURL      string    `json:"banner_url"`
	FollowersCount int       `json:"followers_count"`
	FollowingCount int       `json:"following_count"`
	TweetsCount    int       `json:"tweets_count"`
	Verified       bool      `json:"verified"`
	JoinedAt       time.Time `json:"joined_at"`
}

type Tweet struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Author    Profile   `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Media     []string  `json:"media,omitempty"`
	Metrics   Metrics   `json:"metrics"`
	IsReply   bool      `json:"is_reply"`
	IsRetweet bool      `json:"is_retweet"`
}

func (t *Tweet) CsvHeaders() []string {
	return []string{
		"id", "url", "author_username", "author_name", "text", "created_at",
		"replies", "retweets", "likes", "views", "is_reply", "is_retweet",
	}
}

func (t *Tweet) CsvRow() []string {
	return []string{
		t.ID,
		t.URL,
		t.Author.Username,
		t.Author.Name,
		t.Text,
		t.CreatedAt.Format(time.RFC3339),
		strconv.Itoa(t.Metrics.Replies),
		strconv.Itoa(t.Metrics.Retweets),
		strconv.Itoa(t.Metrics.Likes),
		strconv.Itoa(t.Metrics.Views),
		strconv.FormatBool(t.IsReply),
		strconv.FormatBool(t.IsRetweet),
	}
}
