package tiktok

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"

	"github.com/gosom/google-maps-scraper/internal/jsonparse"
)

func parseTikTokResponse(ctx context.Context, resp *scrapemate.Response, job *TikTokJob) (any, []scrapemate.IJob, error) {
	var videos []*Video
	var nextJobs []scrapemate.IJob

	if resp.Error != nil {
		return videos, nextJobs, resp.Error
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return videos, nextJobs, nil
	}

	// Try __UNIVERSAL_DATA_FOR_REHYDRATION__ first (modern TikTok)
	if data, found := jsonparse.FindScriptJSONByID(doc, "__UNIVERSAL_DATA_FOR_REHYDRATION__"); found {
		videos = extractFromUniversalData(data, job)
		if len(videos) > 0 {
			return videos, nextJobs, nil
		}
	}

	// Fallback: try SIGI_STATE (older TikTok)
	videos = extractFromSigiState(doc, job)
	if len(videos) > 0 {
		return videos, nextJobs, nil
	}

	// Last resort: return empty with page info for debugging
	return videos, nextJobs, nil
}

func extractFromUniversalData(data map[string]any, job *TikTokJob) []*Video {
	var videos []*Video

	// Path: __DEFAULT_SCOPE__ -> webapp.user-detail -> userInfo for profile
	// Path: __DEFAULT_SCOPE__ -> webapp.video-detail -> itemInfo -> itemStruct for single video
	defaultScope := jsonparse.GetMap(data, "__DEFAULT_SCOPE__")
	if defaultScope == nil {
		return videos
	}

	// Try user profile path (list of videos)
	userDetail := jsonparse.GetMap(defaultScope, "webapp.user-detail")
	if userDetail != nil {
		// User info
		userInfo := jsonparse.GetMap(userDetail, "userInfo")
		user := jsonparse.GetMap(userInfo, "user")

		// Item list (videos)
		itemList := jsonparse.GetSlice(userDetail, "itemList")
		for _, item := range itemList {
			if itemMap, ok := item.(map[string]any); ok {
				video := extractVideoFromItem(itemMap, user, job)
				if video != nil {
					videos = append(videos, video)
				}
			}
		}
	}

	// Try single video path
	videoDetail := jsonparse.GetMap(defaultScope, "webapp.video-detail")
	if videoDetail != nil {
		itemInfo := jsonparse.GetMap(videoDetail, "itemInfo")
		itemStruct := jsonparse.GetMap(itemInfo, "itemStruct")
		if itemStruct != nil {
			video := extractVideoFromItem(itemStruct, nil, job)
			if video != nil {
				videos = append(videos, video)
			}
		}
	}

	return videos
}

func extractVideoFromItem(item map[string]any, defaultUser map[string]any, job *TikTokJob) *Video {
	id := jsonparse.GetString(item, "id")
	if id == "" {
		return nil
	}

	// Author info - prefer from item, fall back to default
	author := jsonparse.GetMap(item, "author")
	if author == nil {
		author = defaultUser
	}

	var profile Profile
	if author != nil {
		profile = Profile{
			ID:        jsonparse.GetString(author, "id"),
			Username:  jsonparse.GetString(author, "uniqueId"),
			Name:      jsonparse.GetString(author, "nickname"),
			AvatarURL: jsonparse.GetString(author, "avatarLarger"),
		}
	} else {
		profile = Profile{Username: job.Username}
	}

	// Stats
	stats := jsonparse.GetMap(item, "stats")
	var metrics Metrics
	if stats != nil {
		metrics = Metrics{
			Views:    jsonparse.GetInt(stats, "playCount"),
			Likes:    jsonparse.GetInt(stats, "diggCount"),
			Comments: jsonparse.GetInt(stats, "commentCount"),
			Shares:   jsonparse.GetInt(stats, "shareCount"),
		}
	}

	// Video URL
	videoURL := jsonparse.GetString(item, "video", "playAddr")
	if videoURL == "" {
		// Construct from ID
		videoURL = "https://www.tiktok.com/@" + profile.Username + "/video/" + id
	}

	return &Video{
		ID:          id,
		URL:         videoURL,
		Author:      profile,
		Description: jsonparse.GetString(item, "desc"),
		CreatedAt:   jsonparse.GetTime(item, "createTime"),
		Metrics:     metrics,
	}
}

func extractFromSigiState(doc *goquery.Document, job *TikTokJob) []*Video {
	var videos []*Video

	// Find SIGI_STATE in script tags
	re := regexp.MustCompile(`SIGI_STATE["\s]*=\s*(\{.+?\});`)

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if !strings.Contains(text, "SIGI_STATE") {
			return
		}

		matches := re.FindStringSubmatch(text)
		if len(matches) < 2 {
			return
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
			return
		}

		// Look for ItemModule which contains video data
		itemModule := jsonparse.GetMap(data, "ItemModule")
		if itemModule != nil {
			for _, v := range itemModule {
				if itemMap, ok := v.(map[string]any); ok {
					video := extractVideoFromItem(itemMap, nil, job)
					if video != nil {
						videos = append(videos, video)
					}
				}
			}
		}
	})

	return videos
}
