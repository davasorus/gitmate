package ghapi

import (
	"context"
	"testing"
)

func TestListMilestones(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/milestones",
		`[{"number":1,"title":"v1","state":"open"},{"number":2,"title":"v2","state":"open"}]`))
	ms, err := c.ListMilestones(context.Background(), "o", "r", "open")
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 2 || ms[0].Title != "v1" || ms[0].Number != 1 {
		t.Fatalf("milestones wrong: %+v", ms)
	}
}

func TestSetMilestone(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PATCH", path: "/repos/o/r/issues/5", body: `{"number":5,"milestone":{"number":1}}`}))
	if err := c.SetMilestone(context.Background(), "o", "r", 5, 1); err != nil {
		t.Fatal(err)
	}
}
