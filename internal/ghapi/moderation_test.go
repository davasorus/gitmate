package ghapi

import (
	"context"
	"testing"
)

func TestEditIssueComment(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PATCH", path: "/repos/o/r/issues/comments/99", body: `{"id":99}`}))
	if err := c.EditIssueComment(context.Background(), "o", "r", 99, "updated"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteIssueComment(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "DELETE", path: "/repos/o/r/issues/comments/99", status: 204}))
	if err := c.DeleteIssueComment(context.Background(), "o", "r", 99); err != nil {
		t.Fatal(err)
	}
}

func TestLockUnlockConversation(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "PUT", path: "/repos/o/r/issues/5/lock", status: 204},
		route{method: "DELETE", path: "/repos/o/r/issues/5/lock", status: 204}))
	if err := c.LockConversation(context.Background(), "o", "r", 5, "resolved"); err != nil {
		t.Fatal(err)
	}
	if err := c.UnlockConversation(context.Background(), "o", "r", 5); err != nil {
		t.Fatal(err)
	}
}
