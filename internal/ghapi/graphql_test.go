package ghapi

import (
	"context"
	"net/http"
	"testing"
)

func TestPRDetailGraphQL(t *testing.T) {
	resp := `{"data":{"repository":{"pullRequest":{
		"number":1,"title":"t","state":"OPEN","body":"b",
		"author":{"login":"alice"},
		"labels":{"nodes":[{"name":"bug"}]},
		"assignees":{"nodes":[{"login":"bob"}]},
		"reviews":{"nodes":[{"author":{"login":"carol"},"state":"APPROVED","body":"ok"}]},
		"reviewThreads":{"nodes":[{"id":"T1","isResolved":false,"comments":{"nodes":[{"author":{"login":"dave"},"body":"nit","path":"a.go","line":3}]}}]},
		"commits":{"nodes":[]}
	}}}}`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	})
	c, _ := newGQLTestClient(t, h)
	d, err := c.PRDetailGraphQL(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if d.Number != 1 || d.Author != "alice" || d.Title != "t" {
		t.Errorf("detail wrong: %+v", d)
	}
	if len(d.Labels) != 1 || d.Labels[0] != "bug" {
		t.Errorf("labels wrong: %v", d.Labels)
	}
	if len(d.Reviews) != 1 || d.Reviews[0].State != "APPROVED" {
		t.Errorf("reviews wrong: %+v", d.Reviews)
	}
	if len(d.Threads) != 1 || d.Threads[0].ID != "T1" || len(d.Threads[0].Comments) != 1 {
		t.Errorf("threads wrong: %+v", d.Threads)
	}
}

func TestResolveThread(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"resolveReviewThread":{"thread":{"id":"T1"}}}}`))
	})
	c, _ := newGQLTestClient(t, h)
	if err := c.ResolveThread(context.Background(), "T1"); err != nil {
		t.Fatal(err)
	}
}

func TestUnresolveThread(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"unresolveReviewThread":{"thread":{"id":"T1"}}}}`))
	})
	c, _ := newGQLTestClient(t, h)
	if err := c.UnresolveThread(context.Background(), "T1"); err != nil {
		t.Fatal(err)
	}
}
