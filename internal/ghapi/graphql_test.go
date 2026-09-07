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

func TestPRListGraphQL(t *testing.T) {
	// two PRs: #1 approved with a passing + a failing check; #2 draft, review required, one pending check
	resp := `{"data":{"repository":{"pullRequests":{"nodes":[
		{"number":1,"title":"first","state":"OPEN","isDraft":false,"reviewDecision":"APPROVED",
		 "author":{"login":"alice"},
		 "labels":{"nodes":[{"name":"bug"}]},
		 "commits":{"nodes":[{"commit":{"statusCheckRollup":{"contexts":{"nodes":[
			{"status":"COMPLETED","conclusion":"SUCCESS"},
			{"status":"COMPLETED","conclusion":"FAILURE"}
		 ]}}}}]}},
		{"number":2,"title":"second","state":"OPEN","isDraft":true,"reviewDecision":"REVIEW_REQUIRED",
		 "author":{"login":"bob"},
		 "labels":{"nodes":[]},
		 "commits":{"nodes":[{"commit":{"statusCheckRollup":{"contexts":{"nodes":[
			{"status":"IN_PROGRESS","conclusion":""}
		 ]}}}}]}}
	]}}}}`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	})
	c, _ := newGQLTestClient(t, h)
	list, err := c.PRListGraphQL(context.Background(), "o", "r", "open")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(list))
	}
	// PR #1: approved, 2 checks (1 pass, 1 fail)
	if list[0].Number != 1 || list[0].Author != "alice" || list[0].ReviewDecision != "APPROVED" {
		t.Errorf("pr0 wrong: %+v", list[0])
	}
	if list[0].ChecksTotal != 2 || list[0].ChecksPassed != 1 || list[0].ChecksFailed != 1 {
		t.Errorf("pr0 check rollup wrong: %+v", list[0])
	}
	if len(list[0].Labels) != 1 || list[0].Labels[0] != "bug" {
		t.Errorf("pr0 labels wrong: %v", list[0].Labels)
	}
	// PR #2: draft, review required, 1 pending check
	if !list[1].Draft || list[1].ReviewDecision != "REVIEW_REQUIRED" {
		t.Errorf("pr1 wrong: %+v", list[1])
	}
	if list[1].ChecksTotal != 1 || list[1].ChecksPending != 1 {
		t.Errorf("pr1 check rollup wrong: %+v", list[1])
	}
}

func TestIssueListGraphQL(t *testing.T) {
	resp := `{"data":{"repository":{"issues":{"nodes":[
		{"number":10,"title":"a bug","state":"OPEN","author":{"login":"carol"},
		 "labels":{"nodes":[{"name":"bug"}]},
		 "assignees":{"nodes":[{"login":"dave"}]}},
		{"number":11,"title":"a chore","state":"OPEN","author":{"login":"erin"},
		 "labels":{"nodes":[]},"assignees":{"nodes":[]}}
	]}}}}`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	})
	c, _ := newGQLTestClient(t, h)
	list, err := c.IssueListGraphQL(context.Background(), "o", "r", "open")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(list))
	}
	if list[0].Number != 10 || list[0].Author != "carol" {
		t.Errorf("issue0 wrong: %+v", list[0])
	}
	if len(list[0].Labels) != 1 || list[0].Labels[0] != "bug" {
		t.Errorf("issue0 labels wrong: %v", list[0].Labels)
	}
	if len(list[0].Assignees) != 1 || list[0].Assignees[0] != "dave" {
		t.Errorf("issue0 assignees wrong: %v", list[0].Assignees)
	}
}
