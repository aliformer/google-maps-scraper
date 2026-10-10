package twitter

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"

	"github.com/gosom/google-maps-scraper/internal/jsonparse"
)

func parseTwitterResponse(ctx context.Context, resp *scrapemate.Response, job *TwitterJob) (any, []scrapemate.IJob, error) {
	var tweets []*Tweet
	var nextJobs []scrapemate.IJob

	if resp.Error != nil {
		return tweets, nextJobs, resp.Error
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return tweets, nextJobs, nil
	}

	// Try __NEXT_DATA__ first
	if data, found := jsonparse.FindScriptJSONByID(doc, "__NEXT_DATA__"); found {
		tweets = extractTweetsFromNextData(data, job)
		if len(tweets) > 0 {
			return tweets, nextJobs, nil
		}
	}

	// Search all script tags for tweet data
	tweets = extractTweetsFromScripts(doc, job)
	if len(tweets) > 0 {
		return tweets, nextJobs, nil
	}

	return tweets, nextJobs, nil
}

func extractTweetsFromNextData(data map[string]any, job *TwitterJob) []*Tweet {
	// Find objects with __typename "Tweet" or "TweetWithVisibilityResults"
	tweetObjects := jsonparse.FindByTypename(data, []string{
		"Tweet",
		"TweetWithVisibilityResults",
		"TweetTombstone",
	}, 50)

	return processTweetObjects(tweetObjects, job)
}

func extractTweetsFromScripts(doc *goquery.Document, job *TwitterJob) []*Tweet {
	var tweets []*Tweet

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		// Look for tweet_results or TweetDetail markers
		if !strings.Contains(text, "tweet_results") && !strings.Contains(text, "__typename") {
			return
		}

		// Try to find JSON in the script
		// Twitter often uses: window.__INITIAL_STATE__ = {...}
		jsonStr := extractJSONFromScript(text)
		if jsonStr == "" {
			return
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return
		}

		tweetObjects := jsonparse.FindByTypename(data, []string{
			"Tweet",
			"TweetWithVisibilityResults",
		}, 50)

		tweets = append(tweets, processTweetObjects(tweetObjects, job)...)
	})

	return tweets
}

func extractJSONFromScript(text string) string {
	// Try common patterns
	patterns := []struct {
		prefix string
		suffix string
	}{
		{"window.__INITIAL_STATE__=", ";"},
		{"window.__INITIAL_STATE__ = ", ";"},
		{`"data":`, ""},
	}

	for _, p := range patterns {
		if idx := strings.Index(text, p.prefix); idx >= 0 {
			start := idx + len(p.prefix)
			remaining := text[start:]

			// Find matching brace
			depth := 0
			end := -1
			for i, c := range remaining {
				if c == '{' {
					depth++
				} else if c == '}' {
					depth--
					if depth == 0 {
						end = i + 1
						break
					}
				}
			}

			if end > 0 {
				return remaining[:end]
			}
		}
	}

	// If whole content looks like JSON
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		return text
	}

	return ""
}

func processTweetObjects(objects []map[string]any, job *TwitterJob) []*Tweet {
	var tweets []*Tweet
	seen := make(map[string]bool)

	for _, obj := range objects {
		tweet := extractTweet(obj, job)
		if tweet != nil && tweet.ID != "" && !seen[tweet.ID] {
			seen[tweet.ID] = true
			tweets = append(tweets, tweet)
		}
	}

	return tweets
}

func extractTweet(obj map[string]any, job *TwitterJob) *Tweet {
	// Handle TweetWithVisibilityResults wrapper
	if jsonparse.GetString(obj, "__typename") == "TweetWithVisibilityResults" {
		obj = jsonparse.GetMap(obj, "tweet")
		if obj == nil {
			return nil
		}
	}

	// Get legacy data (contains most tweet info)
	legacy := jsonparse.GetMap(obj, "legacy")
	if legacy == nil {
		return nil
	}

	// Tweet ID
	id := jsonparse.GetString(obj, "rest_id")
	if id == "" {
		id = jsonparse.GetString(legacy, "id_str")
	}
	if id == "" {
		return nil
	}

	// Author info
	core := jsonparse.GetMap(obj, "core")
	userResults := jsonparse.GetMap(core, "user_results")
	userResult := jsonparse.GetMap(userResults, "result")
	userLegacy := jsonparse.GetMap(userResult, "legacy")

	var author Profile
	if userLegacy != nil {
		author = Profile{
			ID:             jsonparse.GetString(userResult, "rest_id"),
			Username:       jsonparse.GetString(userLegacy, "screen_name"),
			Name:           jsonparse.GetString(userLegacy, "name"),
			Bio:            jsonparse.GetString(userLegacy, "description"),
			AvatarURL:      jsonparse.GetString(userLegacy, "profile_image_url_https"),
			BannerURL:      jsonparse.GetString(userLegacy, "profile_banner_url"),
			FollowersCount: jsonparse.GetInt(userLegacy, "followers_count"),
			FollowingCount: jsonparse.GetInt(userLegacy, "friends_count"),
			TweetsCount:    jsonparse.GetInt(userLegacy, "statuses_count"),
			Verified:       jsonparse.GetString(userLegacy, "verified") == "true",
		}
	} else {
		author = Profile{Username: job.Username}
	}

	// Metrics
	metrics := Metrics{
		Likes:     jsonparse.GetInt(legacy, "favorite_count"),
		Retweets:  jsonparse.GetInt(legacy, "retweet_count"),
		Replies:   jsonparse.GetInt(legacy, "reply_count"),
		Bookmarks: jsonparse.GetInt(legacy, "bookmark_count"),
	}

	// Views from separate object
	views := jsonparse.GetMap(obj, "views")
	if views != nil {
		metrics.Views = jsonparse.GetInt(views, "count")
	}

	// Media URLs
	var media []string
	extMedia := jsonparse.GetMap(legacy, "extended_entities")
	if extMedia != nil {
		mediaList := jsonparse.GetSlice(extMedia, "media")
		for _, m := range mediaList {
			if mMap, ok := m.(map[string]any); ok {
				url := jsonparse.GetString(mMap, "media_url_https")
				if url != "" {
					media = append(media, url)
				}
			}
		}
	}

	// Tweet URL
	tweetURL := "https://x.com/" + author.Username + "/status/" + id

	// Detect reply/retweet
	isReply := jsonparse.GetString(legacy, "in_reply_to_status_id_str") != ""
	isRetweet := strings.HasPrefix(jsonparse.GetString(legacy, "full_text"), "RT @")

	return &Tweet{
		ID:        id,
		URL:       tweetURL,
		Author:    author,
		Text:      jsonparse.GetString(legacy, "full_text"),
		CreatedAt: jsonparse.GetTime(legacy, "created_at"),
		Media:     media,
		Metrics:   metrics,
		IsReply:   isReply,
		IsRetweet: isRetweet,
	}
}
