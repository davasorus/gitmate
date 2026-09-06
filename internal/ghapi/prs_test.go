package ghapi

import (
	"context"
	"testing"
)

func TestListPRs(t *testing.T) {
	body := `[
		{"number":1,"title":"first","state":"open","draft":false,
		 "user":{"login":"alice"},"created_at":"2026-01-01T00:00:00Z",
		 "labels":[{"name":"bug"},{"name":"urgent"}]},
		{"number":2,"title":"second","state":"open","draft":true,
		 "user":{"login":"bob"},"created_at":"2026-01-02T00:00:00Z","labels":[]}
	]`
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/pulls", body))
	prs, err := c.ListPRs(context.Background(), "o", "r", "open")
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(prs))
	}
	if prs[0].Number != 1 || prs[0].Title != "first" || prs[0].Author != "alice" {
		t.Errorf("pr0 wrong: %+v", prs[0])
	}
	if len(prs[0].Labels) != 2 || prs[0].Labels[0] != "bug" {
		t.Errorf("pr0 labels wrong: %v", prs[0].Labels)
	}
	if !prs[1].Draft {
		t.Errorf("pr1 should be draft")
	}
}

func TestListIssues(t *testing.T) {
	// Issues API returns PRs too; ListIssues must filter them out (pull_request key).
	body := `[
		{"number":10,"title":"a bug","state":"open","user":{"login":"carol"},
		 "created_at":"2026-01-01T00:00:00Z"},
		{"number":11,"title":"actually a PR","state":"open","user":{"login":"dave"},
		 "created_at":"2026-01-02T00:00:00Z","pull_request":{"url":"x"}}
	]`
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/issues", body))
	issues, err := c.ListIssues(context.Background(), "o", "r", "open")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue (PR filtered out), got %d", len(issues))
	}
	if issues[0].Number != 10 || issues[0].Author != "carol" {
		t.Errorf("issue wrong: %+v", issues[0])
	}
}
