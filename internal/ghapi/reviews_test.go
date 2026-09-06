package ghapi

import (
	"context"
	"testing"
)

func TestListReviews(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/pulls/1/reviews",
		`[{"id":5,"state":"APPROVED","body":"lgtm","user":{"login":"alice"}}]`))
	rvs, err := c.ListReviews(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rvs) != 1 || rvs[0].State != "APPROVED" || rvs[0].Author != "alice" {
		t.Fatalf("reviews wrong: %+v", rvs)
	}
}

func TestSubmitReview(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/pulls/1/reviews", body: `{"id":6,"state":"COMMENTED"}`}))
	err := c.SubmitReview(context.Background(), "o", "r", 1, "COMMENT", "note",
		[]ReviewComment{{Path: "a.go", Line: 3, Body: "hi"}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListRequestedReviewers(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/pulls/1/requested_reviewers",
		`{"users":[{"login":"bob"}],"teams":[]}`))
	rr, err := c.ListRequestedReviewers(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rr) != 1 || rr[0].Login != "bob" {
		t.Fatalf("reviewers wrong: %+v", rr)
	}
}

func TestRequestReviewers(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/pulls/1/requested_reviewers", body: `{"number":1}`}))
	if err := c.RequestReviewers(context.Background(), "o", "r", 1, []string{"bob"}); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveReviewer(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/pulls/1/requested_reviewers", status: 200, body: `{"number":1}`}))
	if err := c.RemoveReviewer(context.Background(), "o", "r", 1, "bob"); err != nil {
		t.Fatal(err)
	}
}

func TestListReviewComments(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/pulls/1/comments",
		`[{"id":7,"path":"a.go","line":3,"body":"nit","user":{"login":"carol"}}]`))
	cs, err := c.ListReviewComments(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 1 || cs[0].Path != "a.go" || cs[0].Author != "carol" {
		t.Fatalf("comments wrong: %+v", cs)
	}
}

func TestListIssueComments(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/issues/1/comments",
		`[{"id":8,"body":"general","user":{"login":"dave"}}]`))
	cs, err := c.ListIssueComments(context.Background(), "o", "r", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 1 || cs[0].Body != "general" {
		t.Fatalf("issue comments wrong: %+v", cs)
	}
}

func TestReplyToReviewComment(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/pulls/1/comments", body: `{"id":9}`}))
	if err := c.ReplyToReviewComment(context.Background(), "o", "r", 1, 7, "reply"); err != nil {
		t.Fatal(err)
	}
}
