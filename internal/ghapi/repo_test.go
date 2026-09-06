package ghapi

import (
	"context"
	"testing"
)

func TestCreateRepo(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/user/repos", body: `{"html_url":"http://x/repo"}`}))
	url, err := c.CreateRepo(context.Background(), "myrepo", "desc", true)
	if err != nil {
		t.Fatal(err)
	}
	if url != "http://x/repo" {
		t.Fatalf("got url %q", url)
	}
}
