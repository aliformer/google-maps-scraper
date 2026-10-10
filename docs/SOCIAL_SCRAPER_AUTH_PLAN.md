# Social Scraper: Interactive Login + JSON Extraction

## Context

User wants to make the existing 4 social scrapers (Twitter/X, TikTok, Threads, Facebook) work properly. Current implementation has:
- Empty cookie files (except Twitter with stale cookies)
- Fragile DOM selectors that don't extract real data
- No metrics, timestamps, or media extraction

Solution: Add interactive browser login (user logs in manually, scraper captures cookies) and rewrite parsers to extract embedded JSON state instead of DOM scraping.

## Approach

### Step 1: Create auth package for interactive browser login

Create `auth/auth.go` with interactive login flow using Playwright.

**New file: `auth/auth.go`**
```go
package auth

type LoginResult struct {
    Platform string
    Cookies  string
    Success  bool
    Error    error
}

func InteractiveLogin(ctx context.Context, platform string) (*LoginResult, error)
```

Flow:
1. Open **visible** (non-headless) Playwright browser
2. Navigate to platform login URL:
   - Twitter: `https://x.com/i/flow/login`
   - TikTok: `https://www.tiktok.com/login`
   - Threads: `https://www.threads.net/login`
   - Facebook: `https://www.facebook.com/login`
3. Display instruction: "Please log in. Press Enter in terminal when done."
4. Wait for user input (stdin)
5. Extract all cookies via `page.Context().Cookies()`
6. Format as semicolon-joined string: `key1=val1; key2=val2`
7. Save to `cookies/{platform}.txt`
8. Close browser

Use `playwright.Run()` with `Headless: false`. No credential storage.

### Step 2: Add auth CLI command

**Edit: `main.go`**

Add new run mode `RunModeAuth = "auth"` to existing constants.

Add flag: `--auth <platform>` triggers interactive login for twitter/tiktok/threads/facebook.

In `main()`, add case for `RunModeAuth`:
```go
case runner.RunModeAuth:
    result, err := auth.InteractiveLogin(ctx, cfg.AuthPlatform)
    if err != nil {
        return err
    }
    fmt.Printf("Cookies saved to cookies/%s.txt\n", cfg.AuthPlatform)
```

### Step 3: Rewrite TikTok parser for JSON extraction

**Edit: `tiktok/process.go`**

Replace `parseTikTokResponse` to extract from `__UNIVERSAL_DATA_FOR_REHYDRATION__` script tag.

```go
func parseTikTokResponse(ctx context.Context, resp *scrapemate.Response, job *TikTokJob) (any, []scrapemate.IJob, error) {
    // 1. Find script#__UNIVERSAL_DATA_FOR_REHYDRATION__ via goquery
    // 2. Parse JSON content
    // 3. Navigate to __DEFAULT_SCOPE__.webapp.video-detail.itemInfo.itemStruct
    //    or for profiles: __DEFAULT_SCOPE__.webapp.user-detail.userInfo
    // 4. Extract: id, desc, createTime, author.uniqueId, author.nickname,
    //    stats.playCount, stats.diggCount, stats.commentCount, stats.shareCount
    // 5. Fallback: try SIGI_STATE regex pattern if rehydration not found
}
```

JSON path for video: `__DEFAULT_SCOPE__ -> webapp -> video-detail -> itemInfo -> itemStruct`
JSON path for profile videos: `__DEFAULT_SCOPE__ -> webapp -> user-detail -> itemList`

Fields to extract from `itemStruct`:
- `id` → Video.ID
- `desc` → Video.Description
- `createTime` (unix) → Video.CreatedAt
- `author.uniqueId` → Video.Author.Username
- `author.nickname` → Video.Author.Name
- `author.avatarLarger` → Video.Author.AvatarURL
- `stats.playCount` → Video.Metrics.Views
- `stats.diggCount` → Video.Metrics.Likes
- `stats.commentCount` → Video.Metrics.Comments
- `stats.shareCount` → Video.Metrics.Shares
- `video.playAddr` → Video.URL (actual video URL)

### Step 4: Rewrite Twitter parser for JSON extraction

**Edit: `twitter/process.go`**

Twitter embeds data in `<script id="__NEXT_DATA__">` or various `<script>` tags with JSON.

```go
func parseTwitterResponse(ctx context.Context, resp *scrapemate.Response, job *TwitterJob) (any, []scrapemate.IJob, error) {
    // 1. Try script#__NEXT_DATA__ first
    // 2. If not found, scan all <script> tags for JSON containing "tweet_results"
    // 3. Walk JSON to find objects with __typename == "Tweet" or "TweetWithVisibilityResults"
    // 4. Extract from result.legacy or result.tweet.legacy:
    //    - id_str, full_text, created_at, user.screen_name, user.name
    //    - favorite_count, retweet_count, reply_count, quote_count, bookmark_count
    //    - extended_entities.media[].media_url_https
}
```

Fields to extract from tweet object:
- `rest_id` or `legacy.id_str` → Tweet.ID
- `legacy.full_text` → Tweet.Text
- `legacy.created_at` (Twitter date format) → Tweet.CreatedAt
- `core.user_results.result.legacy.screen_name` → Tweet.Author.Username
- `core.user_results.result.legacy.name` → Tweet.Author.Name
- `legacy.favorite_count` → Tweet.Metrics.Likes
- `legacy.retweet_count` → Tweet.Metrics.Retweets
- `legacy.reply_count` → Tweet.Metrics.Replies
- `legacy.quote_count` (if present) → (add to Metrics if needed)
- `legacy.bookmark_count` → Tweet.Metrics.Bookmarks
- `views.count` → Tweet.Metrics.Views
- `legacy.extended_entities.media[].media_url_https` → Tweet.Media

Add helper:
```go
func findObjectsByTypename(data any, typenames []string) []map[string]any
```
Recursive walker like Facebook pattern.

### Step 5: Rewrite Threads parser for JSON extraction

**Edit: `threads/process.go`**

Threads embeds GraphQL response in `<script type="application/json">` tags.

```go
func parseThreadsResponse(ctx context.Context, resp *scrapemate.Response, job *ThreadsJob) (any, []scrapemate.IJob, error) {
    // 1. Find all <script type="application/json"> tags
    // 2. Parse each, look for objects with __typename == "XDTThreadItem"
    // 3. Extract thread data from thread_items array
}
```

Fields to extract (path: `data.data.edges[].node.thread_items[]`):
- `post.pk` → Post.ID
- `post.caption.text` → Post.Text
- `post.taken_at` (unix) → Post.CreatedAt
- `post.user.username` → Post.Author.Username
- `post.user.full_name` → Post.Author.Name
- `post.like_count` → Post.Metrics.Likes
- `post.text_post_app_info.share_info.repost_count` → Post.Metrics.Reposts
- `post.text_post_app_info.direct_reply_count` → Post.Metrics.Replies
- `post.carousel_media[].image_versions2.candidates[0].url` → Post.Media

### Step 6: Rewrite Facebook parser for JSON extraction

**Edit: `facebook/process.go`**

Facebook embeds data in `<script type="application/json">` with `__typename` discriminators.

```go
func parseFacebookResponse(ctx context.Context, resp *scrapemate.Response, job *FacebookJob) (any, []scrapemate.IJob, error) {
    // 1. Find all <script type="application/json"> tags
    // 2. Parse each, walk recursively for __typename == "Story" or "Post"
    // 3. Extract post data
}
```

Fields to extract:
- `id` → Post.ID
- `message.text` → Post.Text
- `creation_time` (unix) → Post.CreatedAt
- `actors[0].name` → Post.Author.Name
- `actors[0].id` → Post.Author.ID
- `feedback.reactors.count` → Post.Metrics.Likes
- `feedback.comment_count.total_count` → Post.Metrics.Comments
- `feedback.share_count.count` → Post.Metrics.Shares
- `attachments[].media.image.uri` → Post.Media

### Step 7: Improve BrowserActions wait strategy

**Edit all: `twitter/job.go`, `tiktok/job.go`, `threads/job.go`, `facebook/job.go`**

Replace fixed 2-second wait with dynamic wait for content:

```go
// Before
page.WaitForTimeout(2 * time.Second)

// After
// Try to wait for content indicator, fall back to timeout
_, err := page.WaitForSelector("article, div[data-pressable-container], div[data-e2e]", 
    scrapemate.WaitForSelectorTimeout(5*time.Second))
if err != nil {
    // Content selector not found, wait fixed time as fallback
    page.WaitForTimeout(3 * time.Second)
}
```

Platform-specific selectors:
- Twitter: `article[data-testid='tweet']` or `div[data-testid='cellInnerDiv']`
- TikTok: `div[data-e2e='user-post-item']` or `script#__UNIVERSAL_DATA_FOR_REHYDRATION__`
- Threads: `div[data-pressable-container]` or `script[type='application/json']`
- Facebook: `div[role='article']` or `script[type='application/json']`

### Step 8: Add JSON parsing helpers

**New file: `internal/jsonparse/jsonparse.go`**

Shared utilities for JSON extraction:

```go
package jsonparse

// FindScriptJSON extracts JSON from script tags matching selector
func FindScriptJSON(doc *goquery.Document, selector string) ([]map[string]any, error)

// FindByTypename recursively finds objects with matching __typename
func FindByTypename(data any, typenames []string, maxDepth int) []map[string]any

// GetString safely gets nested string: GetString(obj, "author", "username")
func GetString(obj map[string]any, path ...string) string

// GetInt safely gets nested int
func GetInt(obj map[string]any, path ...string) int

// GetTime parses unix timestamp or date string
func GetTime(obj map[string]any, path ...string) time.Time
```

## Critical Files & Anchors

| File | Symbol/Region | Reason |
|------|---------------|--------|
| `tiktok/process.go:27-45` | `parseTikTokResponse` | DOM selector → JSON extraction rewrite |
| `twitter/process.go:26-43` | `parseTwitterResponse` | DOM selector → JSON extraction rewrite |
| `threads/process.go:26-40` | `parseThreadsResponse` | DOM selector → JSON extraction rewrite |
| `facebook/process.go:26-45` | `parseFacebookResponse` | DOM selector → JSON extraction rewrite |
| `runner/runner.go:20-30` | `RunMode*` constants | Add `RunModeAuth` |

## Verification

### Test interactive login

```bash
cd /mnt/working/sandbox/google-maps-scraper
go build -o bin/scraper .
./bin/scraper --auth twitter
# Browser opens, log in manually, press Enter
cat cookies/twitter.txt  # Should have session cookies
```

### Test TikTok scraping

```bash
# After login:
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{"name":"test","type":"tiktok","data":{"type":"tiktok","username":"tiktok"}}'

# Check output CSV has real metrics (non-zero likes/views)
cat data/*.csv | head -5
```

### Test Twitter scraping

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{"name":"test","type":"twitter","data":{"type":"twitter","username":"elonmusk"}}'

# Verify: tweets have actual timestamps (not time.Now), real metrics
```

### Test all platforms

For each platform (twitter, tiktok, threads, facebook):
1. Run `./bin/scraper --auth {platform}`, complete login
2. Submit scrape job via API
3. Verify CSV output contains:
   - Real post IDs (not synthetic `jobID-0`)
   - Actual timestamps (not current time)
   - Non-zero metrics where posts have engagement
   - Author usernames/names properly separated

## Assumptions & Contingencies

1. **Playwright available**: The auth flow requires Playwright. If `npx playwright install chromium` hasn't been run, the auth command should detect this and print installation instructions.

2. **JSON schema changes**: Social platforms change their JSON structure. If extraction fails (returns 0 results with valid cookies):
   - Log the raw JSON to a debug file
   - Fall back to current DOM extraction as last resort
   - Parser functions should return partial data rather than failing completely

3. **Rate limiting**: If platform returns rate limit response (429 or equivalent):
   - Detect in BrowserActions via response status or page content
   - Return descriptive error rather than garbage fallback data
   - User should wait or rotate cookies

4. **Cookie expiry**: Session cookies expire. If scraping returns login page:
   - Detect login redirect (URL contains /login or /signin)
   - Return clear error: "Session expired. Run --auth {platform} to re-login"
