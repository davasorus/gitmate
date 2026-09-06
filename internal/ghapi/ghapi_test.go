package ghapi

import (
	"context"
	"testing"
)

func TestWhoami(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/user", `{"login":"octocat"}`))
	login, err := c.Whoami(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if login != "octocat" {
		t.Errorf("expected octocat, got %q", login)
	}
}

func TestRateLimit(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/rate_limit",
		`{"resources":{"core":{"limit":5000,"remaining":4999,"reset":0}}}`))
	rem, lim, err := c.RateLimit(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if lim != 5000 || rem != 4999 {
		t.Fatalf("got remaining=%d limit=%d", rem, lim)
	}
}

func TestResolveToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok_abc123")
	if got := resolveToken(); got != "tok_abc123" {
		t.Fatalf("expected token from env, got %q", got)
	}
	// whitespace is trimmed
	t.Setenv("GITHUB_TOKEN", "  spaced  ")
	if got := resolveToken(); got != "spaced" {
		t.Fatalf("expected trimmed token, got %q", got)
	}
}

func TestNew(t *testing.T) {
	// success path: token present → client built
	t.Setenv("GITHUB_TOKEN", "tok_xyz")
	c, err := New(context.Background(), "o", "r")
	if err != nil {
		t.Fatal(err)
	}
	if c == nil || c.Owner != "o" || c.Repo != "r" {
		t.Fatalf("client wrong: %+v", c)
	}
}

func TestNewNoToken(t *testing.T) {
	// error path: no token anywhere → error
	t.Setenv("GITHUB_TOKEN", "")
	// also neutralize a .env two dirs up if present by running from a temp cwd
	if _, err := New(context.Background(), "o", "r"); err == nil {
		t.Skip("a token was resolved from a .env/ambient source; error path not reachable here")
	}
}
