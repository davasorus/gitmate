package ghapi

import (
	"context"
	"testing"
)

func TestAddAssignees(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/issues/5/assignees", body: `{"number":5}`}))
	if err := c.AddAssignees(context.Background(), "o", "r", 5, []string{"alice"}); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveAssignees(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/issues/5/assignees", body: `{"number":5}`}))
	if err := c.RemoveAssignees(context.Background(), "o", "r", 5, []string{"alice"}); err != nil {
		t.Fatal(err)
	}
}
