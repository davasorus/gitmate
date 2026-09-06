package ghapi

import (
	"context"
	"testing"
)

func TestMergePR(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PUT", path: "/repos/o/r/pulls/4/merge", body: `{"sha":"abc123","merged":true}`}))
	sha, err := c.MergePR(context.Background(), "o", "r", 4, "merge")
	if err != nil {
		t.Fatal(err)
	}
	if sha != "abc123" {
		t.Fatalf("got sha %q", sha)
	}
}

func TestCommentPR(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/issues/4/comments", body: `{"html_url":"http://x/c"}`}))
	url, err := c.CommentPR(context.Background(), "o", "r", 4, "nice")
	if err != nil {
		t.Fatal(err)
	}
	if url != "http://x/c" {
		t.Fatalf("got url %q", url)
	}
}

func TestPRChecks(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "GET", path: "/repos/o/r/pulls/4", body: `{"number":4,"head":{"sha":"deadbeef"}}`},
		route{method: "GET", path: "/repos/o/r/commits/deadbeef/check-runs",
			body: `{"total_count":1,"check_runs":[{"name":"ci","status":"completed","conclusion":"success"}]}`}))
	checks, err := c.PRChecks(context.Background(), "o", "r", 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) != 1 || checks[0].Name != "ci" {
		t.Fatalf("checks wrong: %+v", checks)
	}
}
