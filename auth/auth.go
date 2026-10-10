package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mxschmitt/playwright-go"
)

// Platform login URLs
var platformURLs = map[string]string{
	"twitter":  "https://x.com/i/flow/login",
	"tiktok":   "https://www.tiktok.com/login",
	"threads":  "https://www.threads.net/login",
	"facebook": "https://www.facebook.com/login",
}

// LoginResult holds the result of an interactive login
type LoginResult struct {
	Platform string
	Cookies  string
	Success  bool
}

// InteractiveLogin opens a visible browser for the user to log in manually,
// then captures the session cookies and saves them to cookies/{platform}.txt
func InteractiveLogin(ctx context.Context, platform string) (*LoginResult, error) {
	platform = strings.ToLower(platform)
	loginURL, ok := platformURLs[platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s (supported: twitter, tiktok, threads, facebook)", platform)
	}

	// Ensure playwright is installed
	if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
		return nil, fmt.Errorf("failed to install playwright: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch visible browser (headful mode)
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	// Create context and page
	browserCtx, err := browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	page, err := browserCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Navigate to login page
	if _, err := page.Goto(loginURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", loginURL, err)
	}

	// Prompt user
	fmt.Printf("\n=== Interactive Login for %s ===\n", strings.Title(platform))
	fmt.Println("A browser window has opened.")
	fmt.Println("Please log in to your account.")
	fmt.Println("After logging in successfully, press ENTER here to capture cookies.")
	fmt.Print("\nPress ENTER when done... ")

	// Wait for user input
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// Extract cookies
	cookies, err := browserCtx.Cookies()
	if err != nil {
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}

	if len(cookies) == 0 {
		return &LoginResult{Platform: platform, Success: false}, fmt.Errorf("no cookies captured - login may have failed")
	}

	// Format cookies as semicolon-joined string: key1=val1; key2=val2
	var parts []string
	for _, c := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	cookieStr := strings.Join(parts, "; ")

	// Save to cookies/{platform}.txt
	cookieDir := "cookies"
	if err := os.MkdirAll(cookieDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cookies directory: %w", err)
	}

	cookiePath := filepath.Join(cookieDir, platform+".txt")
	if err := os.WriteFile(cookiePath, []byte(cookieStr), 0600); err != nil {
		return nil, fmt.Errorf("failed to write cookies file: %w", err)
	}

	fmt.Printf("\n✓ Captured %d cookies for %s\n", len(cookies), platform)

	return &LoginResult{
		Platform: platform,
		Cookies:  cookieStr,
		Success:  true,
	}, nil
}

// ValidatePlatform checks if a platform name is supported
func ValidatePlatform(platform string) error {
	platform = strings.ToLower(platform)
	if _, ok := platformURLs[platform]; !ok {
		return fmt.Errorf("unsupported platform: %s (supported: twitter, tiktok, threads, facebook)", platform)
	}
	return nil
}
