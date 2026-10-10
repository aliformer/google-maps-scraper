package threads

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosom/scrapemate"

	"github.com/gosom/google-maps-scraper/cookiepool"
	"github.com/gosom/google-maps-scraper/deduper"
	"github.com/gosom/google-maps-scraper/exiter"
)

type ThreadsJobOptions func(*ThreadsJob)

type ThreadsJob struct {
	scrapemate.Job

	Username    string
	MaxDepth    int
	Deduper     deduper.Deduper
	ExitMonitor exiter.Exiter
	Cookie      string
}

func NewThreadsJob(id, username string, maxDepth int, opts ...ThreadsJobOptions) *ThreadsJob {
	if id == "" {
		id = uuid.New().String()
	}

	targetURL := fmt.Sprintf("https://www.threads.net/@%s", strings.TrimPrefix(username, "@"))

	const (
		maxRetries = 3
		prio       = scrapemate.PriorityLow
	)

	job := ThreadsJob{
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
		MaxDepth: maxDepth,
		Cookie:   cookiepool.GetDefaultPool().GetCookie("threads"),
	}

	for _, opt := range opts {
		opt(&job)
	}

	if job.Cookie != "" {
		job.Headers["Cookie"] = job.Cookie
	}

	return &job
}

func WithDeduper(d deduper.Deduper) ThreadsJobOptions {
	return func(j *ThreadsJob) {
		j.Deduper = d
	}
}

func WithExitMonitor(e exiter.Exiter) ThreadsJobOptions {
	return func(j *ThreadsJob) {
		j.ExitMonitor = e
	}
}

func WithCookie(c string) ThreadsJobOptions {
	return func(j *ThreadsJob) {
		if c != "" {
			j.Cookie = c
		}
	}
}

func (j *ThreadsJob) UseInResults() bool {
	return true
}

func (j *ThreadsJob) ProcessOnFetchError() bool {
	return false
}

func (j *ThreadsJob) BrowserActions(ctx context.Context, page scrapemate.BrowserPage) scrapemate.Response {
	var resp scrapemate.Response

	if j.Cookie != "" {
		_, _ = page.Goto("https://www.threads.net/robots.txt", scrapemate.WaitUntilDOMContentLoaded)
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

	// Wait for content or JSON data, fall back to timeout
	err = page.WaitForSelector(`div[data-pressable-container], script[type="application/json"]`, 5*time.Second)
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

func (j *ThreadsJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	// ponytail: basic DOM scraper boilerplate. Upgrade to GraphQL/JSON endpoint parsing.
	return parseThreadsResponse(ctx, resp, j)
}
