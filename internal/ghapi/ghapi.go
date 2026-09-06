package ghapi

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/joho/godotenv"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Client wraps the go-github client plus the resolved owner/repo.
type Client struct {
	gh    *github.Client
	gql   *githubv4.Client
	Owner string
	Repo  string
}

// New builds an authenticated client. The token is resolved from, in order:
//  1. the GITHUB_TOKEN process environment variable
//  2. a .env file (walking up from the working directory)
//  3. the persisted user environment variable (Windows), via `gh auth token`
//
// The first non-empty source wins.
func New(ctx context.Context, owner, repo string) (*Client, error) {
	token := resolveToken()
	if token == "" {
		return nil, errors.New("no GitHub token found — set GITHUB_TOKEN, add it to a .env file, or run `gh auth login`")
	}
	// v88 API (official README): NewClient takes ONLY options and returns
	// (client, error) — no positional http.Client arg. Auth via WithAuthToken.
	gh, err := github.NewClient(github.WithAuthToken(token))
	if err != nil {
		return nil, err
	}
	// shurcooL/githubv4 still takes an *http.Client; build an oauth2 one for it.
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)
	return &Client{
		gh:    gh,
		gql:   githubv4.NewClient(httpClient),
		Owner: owner,
		Repo:  repo,
	}, nil
}

// resolveToken tries several sources and returns the first non-empty token.
func resolveToken() string {
	// 1. Already in the process environment? Use it.
	if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
		return t
	}

	// 2. Load a .env file if present. godotenv.Load walks the current dir;
	//    we also try the module root one level up (the GUI runs from gui/).
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
		return t
	}

	return ""
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
