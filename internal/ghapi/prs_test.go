package ghapi

import (
	"context"
	"net/http"
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
	if len(prs) != 2 || prs[0].Number != 1 || prs[0].Author != "alice" {
		t.Fatalf("prs wrong: %+v", prs)
	}
	if len(prs[0].Labels) != 2 || prs[0].Labels[0] != "bug" {
		t.Errorf("pr0 labels wrong: %v", prs[0].Labels)
	}
	if !prs[1].Draft {
		t.Errorf("pr1 should be draft")
	}
}

func TestListIssues(t *testing.T) {
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
	if len(issues) != 1 || issues[0].Number != 10 || issues[0].Author != "carol" {
		t.Fatalf("issues wrong (PR should be filtered): %+v", issues)
	}
}

func TestPRDiff(t *testing.T) {
	diff := "diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n@@ -1 +1 @@\n-old\n+new\n"
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/o/r/pulls/1" {
			w.Header().Set("Content-Type", "application/vnd.github.v3.diff")
			_, _ = w.Write([]byte(diff))
			return
		}
		w.WriteHeader(404)
	})
	c, _ := newTestClient(t, h)
	files, err := c.PRDiff(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(files))
	}
}
