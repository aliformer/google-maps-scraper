package threads

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
)

func TestParseThreadsResponse(t *testing.T) {
	// JSON embedded in script tag, matching Threads' actual format
	html := `<html><body>
<script type="application/json">
{
  "data": {
    "edges": [{
      "node": {
        "__typename": "XDTThreadItem",
        "thread_items": [{
          "post": {
            "pk": "post123",
            "code": "ABC123",
            "user": {
              "pk": "user456",
              "username": "ThreadsUser",
              "full_name": "Threads User",
              "is_verified": "false"
            },
            "caption": {
              "text": "Hello Threads"
            },
            "taken_at": 1696694400,
            "like_count": 100,
            "text_post_app_info": {
              "direct_reply_count": 5,
              "share_info": {
                "repost_count": 10,
                "quote_count": 3
              }
            }
          }
        }]
      }
    }]
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

	job := NewThreadsJob("test-1", "ThreadsUser", 1)
	res, _, err := parseThreadsResponse(context.Background(), resp, job)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	posts, ok := res.([]*Post)
	if !ok || len(posts) == 0 {
		t.Fatalf("expected posts, got %v", res)
	}

	p := posts[0]
	if p.Text != "Hello Threads" {
		t.Errorf("expected text 'Hello Threads', got %q", p.Text)
	}
	if p.ID != "post123" {
		t.Errorf("expected ID 'post123', got %q", p.ID)
	}
	if p.Author.Username != "ThreadsUser" {
		t.Errorf("expected username 'ThreadsUser', got %q", p.Author.Username)
	}
	if p.Metrics.Likes != 100 {
		t.Errorf("expected 100 likes, got %d", p.Metrics.Likes)
	}
}
