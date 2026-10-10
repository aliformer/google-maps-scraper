package authrunner

import (
	"context"
	"fmt"

	"github.com/gosom/google-maps-scraper/auth"
	"github.com/gosom/google-maps-scraper/runner"
)

type authRunner struct {
	platform string
}

// New creates an auth runner for interactive social media login
func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.RunMode != runner.RunModeAuth {
		return nil, fmt.Errorf("%w: %d", runner.ErrInvalidRunMode, cfg.RunMode)
	}

	if err := auth.ValidatePlatform(cfg.AuthPlatform); err != nil {
		return nil, err
	}

	return &authRunner{platform: cfg.AuthPlatform}, nil
}

func (r *authRunner) Run(ctx context.Context) error {
	result, err := auth.InteractiveLogin(ctx, r.platform)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("login failed for %s", r.platform)
	}

	fmt.Printf("Cookies saved to cookies/%s.txt\n", r.platform)
	return nil
}

func (r *authRunner) Close(context.Context) error {
	return nil
}
