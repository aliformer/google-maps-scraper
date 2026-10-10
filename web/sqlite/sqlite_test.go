package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gosom/google-maps-scraper/web"
)

func TestSqliteCategoryFilter(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "jobs.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	ctx := context.Background()

	jobTwitter := web.Job{
		ID:     "j-tw",
		Name:   "Twitter Scrape",
		Status: web.StatusPending,
		Type:   "twitter",
		Date:   time.Now().UTC(),
		Data: web.JobData{
			Type:     "twitter",
			Username: "elonmusk",
		},
	}

	jobFB := web.Job{
		ID:     "j-fb",
		Name:   "FB Scrape",
		Status: web.StatusPending,
		Type:   "facebook",
		Date:   time.Now().UTC(),
		Data: web.JobData{
			Type:     "facebook",
			Username: "zuck",
		},
	}

	if err := repo.Create(ctx, &jobTwitter); err != nil {
		t.Fatalf("create twitter job failed: %v", err)
	}

	if err := repo.Create(ctx, &jobFB); err != nil {
		t.Fatalf("create fb job failed: %v", err)
	}

	twJobs, err := repo.Select(ctx, web.SelectParams{Type: "twitter"})
	if err != nil {
		t.Fatalf("select twitter failed: %v", err)
	}

	if len(twJobs) != 1 || twJobs[0].ID != "j-tw" {
		t.Fatalf("expected 1 twitter job, got %v", twJobs)
	}

	fbJobs, err := repo.Select(ctx, web.SelectParams{Type: "facebook"})
	if err != nil {
		t.Fatalf("select fb failed: %v", err)
	}

	if len(fbJobs) != 1 || fbJobs[0].ID != "j-fb" {
		t.Fatalf("expected 1 fb job, got %v", fbJobs)
	}
}
