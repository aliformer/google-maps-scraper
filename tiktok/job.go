package tiktok

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosom/scrapemate"

	"github.com/gosom/google-maps-scraper/cookiepool"
	"github.com/gosom/google-maps-scraper/deduper"
	"github.com/gosom/google-maps-scraper/exiter"
)

type TikTokJobOptions func(*TikTokJob)

type TikTokJob struct {
	scrapemate.Job

	Username    string
	Query       string
	MaxDepth    int
	Deduper     deduper.Deduper
	ExitMonitor exiter.Exiter
	Cookie      string
}

func NewTikTokJob(id, username, query string, maxDepth int, opts ...TikTokJobOptions) *TikTokJob {
	if id == "" {
		id = uuid.New().String()
	}

	var targetURL string
	if username != "" {
		targetURL = fmt.Sprintf("https://www.tiktok.com/@%s", strings.TrimPrefix(username, "@"))
	} else if query != "" {
		targetURL = fmt.Sprintf("https://www.tiktok.com/search?q=%s", url.QueryEscape(query))
	} else {
		targetURL = "https://www.tiktok.com/"
	}

	const (
		maxRetries = 3
		prio       = scrapemate.PriorityLow
	)

	job := TikTokJob{
		Job: scrapemate.Job{
			ID:     id,
			Method: http.MethodGet,
			URL:    targetURL,
			Headers: map[string]string{
				"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
				"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
				"Accept-Language":           "en-US,en;q=0.9",
				"Sec-Ch-Ua":                 `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
				"Sec-Ch-Ua-Mobile":          "?0",
				"Sec-Ch-Ua-Platform":        `"Windows"`,
				"Sec-Fetch-Dest":            "document",
				"Sec-Fetch-Mode":            "navigate",
				"Sec-Fetch-Site":            "none",
				"Sec-Fetch-User":            "?1",
				"Upgrade-Insecure-Requests": "1",
			},
			MaxRetries: maxRetries,
			Priority:   prio,
			CheckResponse: func(resp *scrapemate.Response) bool {
				return resp.StatusCode >= 200 && resp.StatusCode < 500
			},
		},
		Username: username,
		Query:    query,
		MaxDepth: maxDepth,
		Cookie:   cookiepool.GetDefaultPool().GetCookie("tiktok"),
	}

	for _, opt := range opts {
		opt(&job)
	}

	if job.Cookie != "" {
		job.Headers["Cookie"] = job.Cookie
	}

	return &job
}

func WithDeduper(d deduper.Deduper) TikTokJobOptions {
	return func(j *TikTokJob) {
		j.Deduper = d
	}
}

func WithExitMonitor(e exiter.Exiter) TikTokJobOptions {
	return func(j *TikTokJob) {
		j.ExitMonitor = e
	}
}

func WithCookie(c string) TikTokJobOptions {
	return func(j *TikTokJob) {
		if c != "" {
			j.Cookie = c
		}
	}
}

func (j *TikTokJob) UseInResults() bool {
	return true
}

func (j *TikTokJob) ProcessOnFetchError() bool {
	return false
}

func (j *TikTokJob) BrowserActions(ctx context.Context, page scrapemate.BrowserPage) scrapemate.Response {
	var resp scrapemate.Response

	if j.Cookie != "" {
		_, _ = page.Goto("https://www.tiktok.com/robots.txt", scrapemate.WaitUntilDOMContentLoaded)
		parts := strings.Split(j.Cookie, ";")
		for _, p := range parts {
			if strings.TrimSpace(p) != "" {
				_, _ = page.Eval(fmt.Sprintf(`document.cookie = %q`, strings.TrimSpace(p)))
			}
		}
	}

	pageResponse, err := page.Goto(j.GetFullURL(), scrapemate.WaitUntilDOMContentLoaded)
	if err != nil {
		resp.Error = err
		return resp
	}

	// Wait for JSON data script or content, fall back to timeout
	err = page.WaitForSelector(`script#__UNIVERSAL_DATA_FOR_REHYDRATION__, div[data-e2e="user-post-item"]`, 5*time.Second)
	if err != nil {
		page.WaitForTimeout(3 * time.Second)
	}

	resp.URL = pageResponse.URL
	resp.StatusCode = pageResponse.StatusCode
	resp.Headers = pageResponse.Headers

	body, err := page.Content()
	if err != nil {
		resp.Error = err
		return resp
	}

	resp.Body = []byte(body)
	return resp
}

func (j *TikTokJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	// ponytail: basic DOM scraper boilerplate. Upgrade to mobile API signature extraction.
	return parseTikTokResponse(ctx, resp, j)
}
