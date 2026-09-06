package ghapi

import (
	"context"
	"net/http"
	"testing"
)

func TestPRDiff(t *testing.T) {
	// GetRaw with Type: Diff sends Accept: application/vnd.github.v3.diff and
	// expects the raw unified diff as the body.
	diff := "diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n@@ -1 +1 @@\n-old\n+new\n"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/o/r/pulls/1" {
			w.Header().Set("Content-Type", "application/vnd.github.v3.diff")
			_, _ = w.Write([]byte(diff))
			return
		}
		w.WriteHeader(404)
	})
	c, _ := newTestClient(t, handler)
	files, err := c.PRDiff(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(files))
	}
}
