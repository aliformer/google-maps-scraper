package twitter

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
)

func TestParseTwitterResponse(t *testing.T) {
	// JSON embedded in script tag, matching Twitter's actual format
	html := `<html><body>
<script id="__NEXT_DATA__" type="application/json">
{
  "props": {
    "pageProps": {
      "timeline": {
        "instructions": [{
          "entries": [{
            "content": {
              "itemContent": {
                "tweet_results": {
                  "result": {
                    "__typename": "Tweet",
                    "rest_id": "123456789",
                    "core": {
                      "user_results": {
                        "result": {
                          "rest_id": "987654321",
                          "legacy": {
                            "screen_name": "TwitterUser",
                            "name": "Twitter User"
                          }
                        }
                      }
                    },
                    "legacy": {
                      "id_str": "123456789",
                      "full_text": "Hello World Tweet",
                      "created_at": "Mon Oct 07 12:00:00 +0000 2024",
                      "favorite_count": 42,
                      "retweet_count": 10,
                      "reply_count": 5
                    },
                    "views": {
                      "count": "1000"
                    }
                  }
                }
              }
            }
          }]
        }]
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

	job := NewTwitterJob("test-1", "TwitterUser", "", 1)
	res, _, err := parseTwitterResponse(context.Background(), resp, job)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	tweets, ok := res.([]*Tweet)
	if !ok || len(tweets) == 0 {
		t.Fatalf("expected tweets, got %v", res)
	}

	tw := tweets[0]
	if tw.Text != "Hello World Tweet" {
		t.Errorf("expected text 'Hello World Tweet', got %q", tw.Text)
	}
	if tw.ID != "123456789" {
		t.Errorf("expected ID '123456789', got %q", tw.ID)
	}
	if tw.Author.Username != "TwitterUser" {
		t.Errorf("expected username 'TwitterUser', got %q", tw.Author.Username)
	}
	if tw.Metrics.Likes != 42 {
		t.Errorf("expected 42 likes, got %d", tw.Metrics.Likes)
	}
}
