package threads

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"

	"github.com/gosom/google-maps-scraper/internal/jsonparse"
)

func parseThreadsResponse(ctx context.Context, resp *scrapemate.Response, job *ThreadsJob) (any, []scrapemate.IJob, error) {
	var posts []*Post
	var nextJobs []scrapemate.IJob

	if resp.Error != nil {
		return posts, nextJobs, resp.Error
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return posts, nextJobs, nil
	}

	// Threads embeds GraphQL data in <script type="application/json"> tags
	posts = extractPostsFromScripts(doc, job)
	if len(posts) > 0 {
		return posts, nextJobs, nil
	}

	return posts, nextJobs, nil
}

func extractPostsFromScripts(doc *goquery.Document, job *ThreadsJob) []*Post {
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

		// Look for XDTThreadItem or Thread objects
		threadItems := jsonparse.FindByTypename(data, []string{
			"XDTThreadItem",
			"XDTThread",
			"Thread",
		}, 50)

		for _, item := range threadItems {
			post := extractPostFromThreadItem(item, job)
			if post != nil {
				posts = append(posts, post)
			}
		}

		// Also look for posts directly (different structure)
		postObjects := jsonparse.FindByKey(data, "thread_items", 30)
		for _, obj := range postObjects {
			threadItems := jsonparse.GetSlice(obj, "thread_items")
			for _, ti := range threadItems {
				if tiMap, ok := ti.(map[string]any); ok {
					post := extractPostFromThreadItem(tiMap, job)
					if post != nil {
						posts = append(posts, post)
					}
				}
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

func extractPostFromThreadItem(item map[string]any, job *ThreadsJob) *Post {
	// Try to get post data from nested "post" key or directly
	postData := jsonparse.GetMap(item, "post")
	if postData == nil {
		postData = item
	}

	// ID can be pk or id
	id := jsonparse.GetString(postData, "pk")
	if id == "" {
		id = jsonparse.GetString(postData, "id")
	}
	if id == "" {
		return nil
	}

	// Author info
	user := jsonparse.GetMap(postData, "user")
	var author Profile
	if user != nil {
		author = Profile{
			ID:             jsonparse.GetString(user, "pk"),
			Username:       jsonparse.GetString(user, "username"),
			Name:           jsonparse.GetString(user, "full_name"),
			Bio:            jsonparse.GetString(user, "biography"),
			AvatarURL:      jsonparse.GetString(user, "profile_pic_url"),
			FollowersCount: jsonparse.GetInt(user, "follower_count"),
			Verified:       jsonparse.GetString(user, "is_verified") == "true",
		}
	} else {
		author = Profile{Username: job.Username}
	}

	// Text content
	text := ""
	caption := jsonparse.GetMap(postData, "caption")
	if caption != nil {
		text = jsonparse.GetString(caption, "text")
	}
	if text == "" {
		text = jsonparse.GetString(postData, "text")
	}

	// Metrics
	metrics := Metrics{
		Likes: jsonparse.GetInt(postData, "like_count"),
	}

	// Reply/repost counts from text_post_app_info
	textPostInfo := jsonparse.GetMap(postData, "text_post_app_info")
	if textPostInfo != nil {
		metrics.Replies = jsonparse.GetInt(textPostInfo, "direct_reply_count")

		shareInfo := jsonparse.GetMap(textPostInfo, "share_info")
		if shareInfo != nil {
			metrics.Reposts = jsonparse.GetInt(shareInfo, "repost_count")
			metrics.Shares = jsonparse.GetInt(shareInfo, "quote_count")
		}
	}

	// Media URLs
	var media []string

	// Single image
	imageVersions := jsonparse.GetMap(postData, "image_versions2")
	if imageVersions != nil {
		candidates := jsonparse.GetSlice(imageVersions, "candidates")
		if len(candidates) > 0 {
			if first, ok := candidates[0].(map[string]any); ok {
				if url := jsonparse.GetString(first, "url"); url != "" {
					media = append(media, url)
				}
			}
		}
	}

	// Carousel media
	carouselMedia := jsonparse.GetSlice(postData, "carousel_media")
	for _, cm := range carouselMedia {
		if cmMap, ok := cm.(map[string]any); ok {
			iv := jsonparse.GetMap(cmMap, "image_versions2")
			if iv != nil {
				candidates := jsonparse.GetSlice(iv, "candidates")
				if len(candidates) > 0 {
					if first, ok := candidates[0].(map[string]any); ok {
						if url := jsonparse.GetString(first, "url"); url != "" {
							media = append(media, url)
						}
					}
				}
			}
		}
	}

	// Post URL
	code := jsonparse.GetString(postData, "code")
	postURL := "https://www.threads.net/@" + author.Username + "/post/" + code
	if code == "" {
		postURL = "https://www.threads.net/@" + author.Username
	}

	return &Post{
		ID:        id,
		URL:       postURL,
		Author:    author,
		Text:      text,
		CreatedAt: jsonparse.GetTime(postData, "taken_at"),
		Media:     media,
		Metrics:   metrics,
	}
}
