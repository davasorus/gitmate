package ghapi

import (
	"context"
	"testing"
)

func TestParseRepo(t *testing.T) {
	cases := []struct {
		url, owner, repo string
		wantErr          bool
	}{
		{"https://github.com/davasorus/gitmate.git", "davasorus", "gitmate", false},
		{"https://github.com/o/r", "o", "r", false},
		{"git@github.com:o/r.git", "o", "r", false},
		{"https://gitlab.com/o/r", "", "", true},
	}
	for _, tc := range cases {
		o, r, err := ParseRepo(tc.url)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: expected error", tc.url)
			}
			continue
		}
		if err != nil || o != tc.owner || r != tc.repo {
			t.Errorf("%s: got (%q,%q,%v)", tc.url, o, r, err)
		}
	}
}

func TestCreateRepo(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/user/repos", body: `{"clone_url":"http://x/repo.git"}`}))
	url, err := c.CreateRepo(context.Background(), "myrepo", "desc", true)
	if err != nil {
		t.Fatal(err)
	}
	if url != "http://x/repo.git" {
		t.Fatalf("got url %q", url)
	}
}
