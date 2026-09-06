package ghapi

import (
	"context"
	"testing"
)

func TestListLabels(t *testing.T) {
	body := `[{"name":"bug","color":"ff0000","description":"a bug"},{"name":"feat","color":"00ff00","description":""}]`
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/labels", body))
	labels, err := c.ListLabels(context.Background(), "o", "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 2 || labels[0].Name != "bug" || labels[0].Color != "ff0000" {
		t.Fatalf("labels wrong: %+v", labels)
	}
}

func TestCreateLabel(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/labels", body: `{"name":"bug","color":"ff0000"}`}))
	if err := c.CreateLabel(context.Background(), "o", "r", "bug", "ff0000", "desc"); err != nil {
		t.Fatal(err)
	}
}

func TestEditLabel(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PATCH", path: "/repos/o/r/labels/old", body: `{"name":"new","color":"00ff00"}`}))
	if err := c.EditLabel(context.Background(), "o", "r", "old", "new", "00ff00", "d"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteLabel(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/labels/bug", status: 204}))
	if err := c.DeleteLabel(context.Background(), "o", "r", "bug"); err != nil {
		t.Fatal(err)
	}
}

func TestAddLabels(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/issues/5/labels", body: `[{"name":"bug"}]`}))
	if err := c.AddLabels(context.Background(), "o", "r", 5, []string{"bug"}); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveLabel(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/issues/5/labels/bug", body: `[]`}))
	if err := c.RemoveLabel(context.Background(), "o", "r", 5, "bug"); err != nil {
		t.Fatal(err)
	}
}

func TestSetLabels(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PUT", path: "/repos/o/r/issues/5/labels", body: `[{"name":"bug"}]`}))
	if err := c.SetLabels(context.Background(), "o", "r", 5, []string{"bug"}); err != nil {
		t.Fatal(err)
	}
}
