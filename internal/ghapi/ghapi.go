package ghapi

import (
	"context"
	"errors"
	"os"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// Client wraps the go-github client plus the resolved owner/repo.
type Client struct {
	gh    *github.Client
	Owner string
	Repo  string
}

// New builds an authenticated client. The token is read from the
// GITHUB_TOKEN environment variable — same convention as the Wails
// updater tutorial, and what GitHub Actions injects automatically.
func New(ctx context.Context, owner, repo string) (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN not set — create a PAT and set it in your environment")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)
	return &Client{
		gh:    github.NewClient(httpClient),
		Owner: owner,
		Repo:  repo,
	}, nil
}

// Whoami returns the authenticated user's login — a cheap call to
// confirm the token works and to see your rate-limit headers.
func (c *Client) Whoami(ctx context.Context) (string, error) {
	user, _, err := c.gh.Users.Get(ctx, "") // "" means the authenticated user
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// RateLimit reports remaining core API calls — useful to watch while
// you learn, since unauthenticated requests are capped at 60/hour.
func (c *Client) RateLimit(ctx context.Context) (remaining, limit int, err error) {
	rl, _, err := c.gh.RateLimit.Get(ctx)
	if err != nil {
		return 0, 0, err
	}
	return rl.Core.Remaining, rl.Core.Limit, nil
}
