package ghapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/shurcooL/githubv4"
)

// newTestClient returns a *Client whose REST client points at an httptest server
// running the given handler. WithEnterpriseURLs makes go-github prefix requests
// with /api/v3 (and /api/uploads for uploads); we strip those so tests can
// register plain paths like "/repos/o/r/pulls".
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	stripped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/v3")
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/uploads")
		handler.ServeHTTP(w, r)
	})
	srv := httptest.NewServer(stripped)
	t.Cleanup(srv.Close)

	gh, err := github.NewClient(github.WithEnterpriseURLs(srv.URL+"/", srv.URL+"/"))
	if err != nil {
		t.Fatalf("github.NewClient: %v", err)
	}
	return &Client{gh: gh, Owner: "o", Repo: "r"}, srv
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

// jsonHandler routes a single path to a JSON body.
func jsonHandler(t *testing.T, path, body string) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	return mux
}

// route is one mock endpoint: match method+path, return status+body.
type route struct {
	method string
	path   string
	status int
	body   string
}

// routeHandler builds a handler from a set of routes. Unmatched → 404.
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

// repoCommon: shared little helpers can live here as needed.
