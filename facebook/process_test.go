package facebook

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
)

func TestParseFacebookResponse(t *testing.T) {
	// JSON embedded in script tag, matching Facebook's actual format
	html := `<html><body>
<script type="application/json">
{
  "data": {
    "node": {
      "__typename": "Story",
      "id": "post123",
      "url": "https://www.facebook.com/post123",
      "creation_time": 1696694400,
      "actors": [{
        "id": "user456",
        "name": "AuthorName",
        "username": "fbuser"
      }],
      "message": {
        "text": "Sample FB post"
      },
      "feedback": {
        "reactors": {
          "count": 50
        },
        "comment_count": {
          "total_count": 10
        },
        "share_count": {
          "count": 5
        }
      }
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

	job := NewFacebookJob("test-1", "fbuser", "", 1)
	res, _, err := parseFacebookResponse(context.Background(), resp, job)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	posts, ok := res.([]*Post)
	if !ok || len(posts) == 0 {
		t.Fatalf("expected posts, got %v", res)
	}

	p := posts[0]
	if p.Text != "Sample FB post" {
		t.Errorf("expected text 'Sample FB post', got %q", p.Text)
	}
	if p.ID != "post123" {
		t.Errorf("expected ID 'post123', got %q", p.ID)
	}
	if p.Author.Name != "AuthorName" {
		t.Errorf("expected author name 'AuthorName', got %q", p.Author.Name)
	}
	if p.Metrics.Likes != 50 {
		t.Errorf("expected 50 likes, got %d", p.Metrics.Likes)
	}
}
