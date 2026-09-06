package ghapi

import (
	"context"
	"net/http"
	"testing"
)

func TestUploadAsset(t *testing.T) {
	// UploadReleaseAsset POSTs to the upload URL for the release's assets.
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/releases/2/assets",
			body: `{"id":50,"name":"bin.zip","size":11,"browser_download_url":"http://x/bin.zip"}`}))
	asset, err := c.UploadAsset(context.Background(), "o", "r", 2, "bin.zip", []byte("hello-world"))
	if err != nil {
		t.Fatal(err)
	}
	if asset.ID != 50 || asset.Name != "bin.zip" {
		t.Fatalf("asset wrong: %+v", asset)
	}
}

func TestJobLogs(t *testing.T) {
	// GetWorkflowJobLogs hits .../logs and gets a 302 to the log text; go-github
	// returns that Location URL, then JobLogs GETs it. Point both at our server.
	logText := "##[group]Setup\nprep\n##[endgroup]\n##[group]Run\nbuilding\n##[endgroup]\n"
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/o/r/actions/jobs/9/logs":
			http.Redirect(w, r, "http://"+r.Host+"/logblob", http.StatusFound)
		case r.URL.Path == "/logblob":
			_, _ = w.Write([]byte(logText))
		default:
			w.WriteHeader(404)
		}
	})
	c, _ := newTestClient(t, h)
	log, err := c.JobLogs(context.Background(), "o", "r", 9)
	if err != nil {
		t.Fatal(err)
	}
	if len(log.Steps) != 2 || log.Steps[0].Name != "Setup" {
		t.Fatalf("log steps wrong: %+v", log.Steps)
	}
	if log.Raw == "" {
		t.Errorf("expected raw log preserved")
	}
}
