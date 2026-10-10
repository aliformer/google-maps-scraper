package facebook

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"

	"github.com/gosom/google-maps-scraper/internal/jsonparse"
)

func parseFacebookResponse(ctx context.Context, resp *scrapemate.Response, job *FacebookJob) (any, []scrapemate.IJob, error) {
	var posts []*Post
	var nextJobs []scrapemate.IJob

	if resp.Error != nil {
		return posts, nextJobs, resp.Error
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return posts, nextJobs, nil
	}

	// Facebook embeds data in <script type="application/json"> tags with __typename discriminators
	posts = extractPostsFromScripts(doc, job)
	if len(posts) > 0 {
		return posts, nextJobs, nil
	}

	return posts, nextJobs, nil
}

func extractPostsFromScripts(doc *goquery.Document, job *FacebookJob) []*Post {
	var posts []*Post

	doc.Find(`script[type="application/json"]`).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(text), &data); err != nil {
			return
		}

		// Look for Story or Post objects (Facebook's internal types)
		storyObjects := jsonparse.FindByTypename(data, []string{
			"Story",
			"Post",
			"UserPost",
			"PagePost",
		}, 50)

		for _, obj := range storyObjects {
			post := extractPostFromStory(obj, job)
			if post != nil {
				posts = append(posts, post)
			}
		}
	})

	// Also search in regular script tags for embedded JSON
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if !strings.Contains(text, `"__typename"`) {
			return
		}

		// Try to extract JSON from script content
		jsonStr := extractJSONFromScript(text)
		if jsonStr == "" {
			return
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return
		}

		storyObjects := jsonparse.FindByTypename(data, []string{
			"Story",
			"Post",
			"UserPost",
			"PagePost",
		}, 50)

		for _, obj := range storyObjects {
			post := extractPostFromStory(obj, job)
			if post != nil {
				posts = append(posts, post)
			}
		}
	})

	// Deduplicate by ID
	seen := make(map[string]bool)
	var unique []*Post
	for _, p := range posts {
		if p.ID != "" && !seen[p.ID] {
			seen[p.ID] = true
			unique = append(unique, p)
		}
	}

	return unique
}

func extractJSONFromScript(text string) string {
	// Find JSON object boundaries
	text = strings.TrimSpace(text)

	// Look for assignment patterns
	patterns := []string{
		"require(\"ServerJS\").handle(",
		"bigPipe.onPageletArrive(",
		"new (require(\"ServerJS\"))",
	}

	for _, p := range patterns {
		if idx := strings.Index(text, p); idx >= 0 {
			start := idx + len(p)
			remaining := text[start:]

			// Find the JSON object
			braceIdx := strings.Index(remaining, "{")
			if braceIdx >= 0 {
				remaining = remaining[braceIdx:]
				if end := findMatchingBrace(remaining); end > 0 {
					return remaining[:end]
				}
			}
		}
	}

	// Direct JSON
	if strings.HasPrefix(text, "{") {
		if end := findMatchingBrace(text); end > 0 {
			return text[:end]
		}
	}

	return ""
}

func findMatchingBrace(s string) int {
	depth := 0
	for i, c := range s {
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

func extractPostFromStory(obj map[string]any, job *FacebookJob) *Post {
	// ID
	id := jsonparse.GetString(obj, "id")
	if id == "" {
		id = jsonparse.GetString(obj, "post_id")
	}
	if id == "" {
		return nil
	}

	// Author from actors array or comet_sections
	var author Profile

	// Try actors array first
	actors := jsonparse.GetSlice(obj, "actors")
	if len(actors) > 0 {
		if actorMap, ok := actors[0].(map[string]any); ok {
			author = Profile{
				ID:        jsonparse.GetString(actorMap, "id"),
				Name:      jsonparse.GetString(actorMap, "name"),
				Username:  jsonparse.GetString(actorMap, "username"),
				URL:       jsonparse.GetString(actorMap, "url"),
				AvatarURL: jsonparse.GetString(actorMap, "profile_picture", "uri"),
			}
		}
	}

	// Try owner
	if author.ID == "" {
		owner := jsonparse.GetMap(obj, "owner")
		if owner != nil {
			author = Profile{
				ID:        jsonparse.GetString(owner, "id"),
				Name:      jsonparse.GetString(owner, "name"),
				Username:  jsonparse.GetString(owner, "username"),
				URL:       jsonparse.GetString(owner, "url"),
				AvatarURL: jsonparse.GetString(owner, "profile_picture", "uri"),
			}
		}
	}

	if author.ID == "" {
		author = Profile{Username: job.Username}
	}

	// Text content
	text := ""

	// Try message object
	message := jsonparse.GetMap(obj, "message")
	if message != nil {
		text = jsonparse.GetString(message, "text")
	}

	// Try comet_sections path
	if text == "" {
		cometSections := jsonparse.GetMap(obj, "comet_sections")
		if cometSections != nil {
			content := jsonparse.GetMap(cometSections, "content")
			story := jsonparse.GetMap(content, "story")
			cometMsg := jsonparse.GetMap(story, "comet_sections", "message")
			if cometMsg != nil {
				msgStory := jsonparse.GetMap(cometMsg, "story")
				msgObj := jsonparse.GetMap(msgStory, "message")
				text = jsonparse.GetString(msgObj, "text")
			}
		}
	}

	// Metrics from feedback object
	var metrics Metrics
	feedback := jsonparse.GetMap(obj, "feedback")
	if feedback != nil {
		// Reactions/likes
		reactors := jsonparse.GetMap(feedback, "reactors")
		if reactors != nil {
			metrics.Likes = jsonparse.GetInt(reactors, "count")
		}
		if metrics.Likes == 0 {
			metrics.Likes = jsonparse.GetInt(feedback, "reaction_count", "count")
		}

		// Comments
		commentCount := jsonparse.GetMap(feedback, "comment_count")
		if commentCount != nil {
			metrics.Comments = jsonparse.GetInt(commentCount, "total_count")
		}

		// Shares
		shareCount := jsonparse.GetMap(feedback, "share_count")
		if shareCount != nil {
			metrics.Shares = jsonparse.GetInt(shareCount, "count")
		}
	}

	// Media URLs
	var media []string
	attachments := jsonparse.GetSlice(obj, "attachments")
	for _, att := range attachments {
		if attMap, ok := att.(map[string]any); ok {
			mediaObj := jsonparse.GetMap(attMap, "media")
			if mediaObj != nil {
				// Image
				image := jsonparse.GetMap(mediaObj, "image")
				if image != nil {
					if uri := jsonparse.GetString(image, "uri"); uri != "" {
						media = append(media, uri)
					}
				}
				// Photo image
				if uri := jsonparse.GetString(mediaObj, "photo_image", "uri"); uri != "" {
					media = append(media, uri)
				}
			}
		}
	}

	// Post URL
	postURL := jsonparse.GetString(obj, "url")
	if postURL == "" {
		postURL = "https://www.facebook.com/" + id
	}

	// Timestamp
	createdAt := jsonparse.GetTime(obj, "creation_time")
	if createdAt.IsZero() {
		createdAt = jsonparse.GetTime(obj, "created_time")
	}

	return &Post{
		ID:        id,
		URL:       postURL,
		Author:    author,
		Text:      text,
		CreatedAt: createdAt,
		Media:     media,
		Metrics:   metrics,
	}
}
