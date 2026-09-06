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
