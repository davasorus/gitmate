package ghapi

import (
	"context"
	"testing"
)

func TestListReleases(t *testing.T) {
	body := `[{"id":1,"tag_name":"v1.0","name":"one","body":"notes","draft":false,"prerelease":false,"immutable":true,"html_url":"http://x/1"}]`
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/releases", body))
	rels, err := c.ListReleases(context.Background(), "o", "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 1 || rels[0].TagName != "v1.0" || !rels[0].Immutable {
		t.Fatalf("releases wrong: %+v", rels)
	}
}

func TestCreateRelease(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/releases", body: `{"id":2,"tag_name":"v2.0","name":"two"}`}))
	rel, err := c.CreateRelease(context.Background(), "o", "r", "v2.0", "two", "b", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if rel.TagName != "v2.0" {
		t.Fatalf("got %+v", rel)
	}
}

func TestEditRelease(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PATCH", path: "/repos/o/r/releases/2", body: `{"id":2,"name":"edited"}`}))
	rel, err := c.EditRelease(context.Background(), "o", "r", 2, "edited", "b", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if rel.Name != "edited" {
		t.Fatalf("got %+v", rel)
	}
}

func TestDeleteRelease(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/releases/2", status: 204}))
	if err := c.DeleteRelease(context.Background(), "o", "r", 2); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateReleaseNotes(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/releases/generate-notes", body: `{"name":"v3 notes","body":"changelog"}`}))
	name, body, err := c.GenerateReleaseNotes(context.Background(), "o", "r", "v3.0")
	if err != nil {
		t.Fatal(err)
	}
	if name != "v3 notes" || body != "changelog" {
		t.Fatalf("got name=%q body=%q", name, body)
	}
}

func TestListAssets(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/releases/2/assets",
		`[{"id":10,"name":"bin.zip","size":123,"browser_download_url":"http://x/bin.zip"}]`))
	assets, err := c.ListAssets(context.Background(), "o", "r", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].Name != "bin.zip" || assets[0].Size != 123 {
		t.Fatalf("assets wrong: %+v", assets)
	}
}

func TestDeleteAsset(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/releases/assets/10", status: 204}))
	if err := c.DeleteAsset(context.Background(), "o", "r", 10); err != nil {
		t.Fatal(err)
	}
}

func TestDownloadAsset(t *testing.T) {
	// DownloadAsset fetches the asset (GET on the asset endpoint) then its content.
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "GET", path: "/repos/o/r/releases/assets/10", body: `hello-bytes`}))
	data, err := c.DownloadAsset(context.Background(), "o", "r", 10)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello-bytes" {
		t.Fatalf("got %q", string(data))
	}
}
