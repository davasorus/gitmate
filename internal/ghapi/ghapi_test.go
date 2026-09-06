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
