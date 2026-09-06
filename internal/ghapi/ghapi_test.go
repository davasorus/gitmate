package ghapi

import (
	"context"
	"os"
	"path/filepath"
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
	t.Setenv("GITHUB_TOKEN", "  spaced  ")
	if got := resolveToken(); got != "spaced" {
		t.Fatalf("expected trimmed token, got %q", got)
	}
}

func TestNew(t *testing.T) {
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
	// Force the no-token path deterministically: empty env AND a cwd with no .env
	// to find (so godotenv.Load() and ../.env both come up empty in CI and locally).
	t.Setenv("GITHUB_TOKEN", "")
	empty := t.TempDir()
	sub := filepath.Join(empty, "sub") // two levels so ../.env also misses
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	oldwd, _ := os.Getwd()
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	if _, err := New(context.Background(), "o", "r"); err == nil {
		t.Fatal("expected error when no token is resolvable")
	}
}
