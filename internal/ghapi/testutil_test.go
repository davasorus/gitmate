package ghapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/shurcooL/githubv4"
)

// newTestClient spins up an httptest server with the given handler and returns a
// *Client whose REST client points at it — so ghapi methods hit the mock instead
// of the real GitHub API. Fields are unexported, so this lives in-package.
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	// WithEnterpriseURLs makes go-github prefix requests with /api/v3; strip it so
	// tests register plain paths like "/repos/o/r/pulls".
	stripped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/v3")
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/uploads")
		handler.ServeHTTP(w, r)
	})
	srv := httptest.NewServer(stripped)
	t.Cleanup(srv.Close)

	// v88: base/upload URLs are set via a ClientOption, not field assignment.
	// WithEnterpriseURLs points both the API and upload base at our mock server.
	gh, err := github.NewClient(github.WithEnterpriseURLs(srv.URL+"/", srv.URL+"/"))
	if err != nil {
		t.Fatalf("github.NewClient: %v", err)
	}
	return &Client{gh: gh, Owner: "o", Repo: "r"}, srv
}

// mux is a tiny helper: route by exact path to a JSON responder.
func jsonHandler(t *testing.T, path, body string) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	return mux
}

// route is one mock endpoint: match method+path, return status+body, and optionally
// capture the request body for assertions.
type route struct {
	method string
	path   string
	status int
	body   string
}

// routeHandler builds a handler from a set of routes. Unmatched → 404.
// If a route's status is 0 it defaults to 200.
func routeHandler(t *testing.T, routes ...route) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, rt := range routes {
			if rt.method != "" && rt.method != r.Method {
				continue
			}
			if rt.path != r.URL.Path {
				continue
			}
			st := rt.status
			if st == 0 {
				st = 200
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(st)
			_, _ = w.Write([]byte(rt.body))
			return
		}
		w.WriteHeader(404)
	})
}

// newGQLTestClient returns a *Client whose GraphQL client points at an httptest
// server. The handler should respond to POST / with a {"data":...} JSON body.
func newGQLTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	gql := githubv4.NewEnterpriseClient(srv.URL, srv.Client())
	return &Client{gql: gql, Owner: "o", Repo: "r"}, srv
}
