package tiktok

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
)

func TestParseTikTokResponse(t *testing.T) {
	// JSON embedded in script tag, matching TikTok's actual format
	html := `<html><body>
<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">
{
  "__DEFAULT_SCOPE__": {
    "webapp.user-detail": {
      "userInfo": {
        "user": {
          "id": "user123",
          "uniqueId": "ttuser",
          "nickname": "TikTok User"
        }
      },
      "itemList": [{
        "id": "video123",
        "desc": "Cool TikTok video",
        "createTime": 1696694400,
        "author": {
          "id": "user123",
          "uniqueId": "ttuser",
          "nickname": "TikTok User",
          "avatarLarger": "https://example.com/avatar.jpg"
        },
        "stats": {
          "playCount": 10000,
          "diggCount": 500,
          "commentCount": 50,
          "shareCount": 25
        },
        "video": {
          "playAddr": "https://example.com/video.mp4"
        }
      }]
    }
  }
}
</script>
</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := &scrapemate.Response{
		Document: doc,
	}

	job := NewTikTokJob("test-1", "ttuser", "", 1)
	res, _, err := parseTikTokResponse(context.Background(), resp, job)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	videos, ok := res.([]*Video)
	if !ok || len(videos) == 0 {
		t.Fatalf("expected videos, got %v", res)
	}

	v := videos[0]
	if v.Description != "Cool TikTok video" {
		t.Errorf("expected description 'Cool TikTok video', got %q", v.Description)
	}
	if v.ID != "video123" {
		t.Errorf("expected ID 'video123', got %q", v.ID)
	}
	if v.Author.Username != "ttuser" {
		t.Errorf("expected username 'ttuser', got %q", v.Author.Username)
	}
	if v.Metrics.Views != 10000 {
		t.Errorf("expected 10000 views, got %d", v.Metrics.Views)
	}
	if v.Metrics.Likes != 500 {
		t.Errorf("expected 500 likes, got %d", v.Metrics.Likes)
	}
}
