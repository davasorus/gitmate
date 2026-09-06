package ghapi

import (
	"context"
	"testing"
)

func TestRateLimit(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/rate_limit",
		`{"resources":{"core":{"limit":5000,"remaining":4999,"reset":0}}}`))
	rem, lim, err := c.RateLimit(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if lim != 5000 || rem != 4999 {
		t.Fatalf("got remaining=%d limit=%d", rem, lim)
	}
}
