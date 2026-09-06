package ghapi

import (
	"context"
	"testing"
)

func TestCreatePR(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/pulls", body: `{"number":7,"html_url":"http://x/7"}`}))
	n, url, err := c.CreatePR(context.Background(), "o", "r", "t", "b", "head", "base")
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 || url != "http://x/7" {
		t.Fatalf("got n=%d url=%q", n, url)
	}
}

func TestCreateIssue(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/issues", body: `{"number":9,"html_url":"http://x/9"}`}))
	n, url, err := c.CreateIssue(context.Background(), "o", "r", "t", "b")
	if err != nil {
		t.Fatal(err)
	}
	if n != 9 || url != "http://x/9" {
		t.Fatalf("got n=%d url=%q", n, url)
	}
}

func TestSetIssueState(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PATCH", path: "/repos/o/r/issues/3", body: `{"number":3,"state":"closed"}`}))
	if err := c.SetIssueState(context.Background(), "o", "r", 3, "closed"); err != nil {
		t.Fatal(err)
	}
}

func TestSetPRState(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PATCH", path: "/repos/o/r/pulls/3", body: `{"number":3,"state":"closed"}`}))
	if err := c.SetPRState(context.Background(), "o", "r", 3, "closed"); err != nil {
		t.Fatal(err)
	}
}
