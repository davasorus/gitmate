package ghapi

import (
	"context"
	"encoding/base64"
	"testing"
)

// contentJSON builds a GitHub "get contents" response with base64-encoded body.
func contentJSON(body string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(body))
	return `{"type":"file","encoding":"base64","content":"` + enc + `","name":"ci.yml","path":".github/workflows/ci.yml"}`
}

func TestListDispatchableWorkflows(t *testing.T) {
	wfList := `{"total_count":1,"workflows":[{"id":1,"name":"CI","state":"active","path":".github/workflows/ci.yml"}]}`
	yml := "on:\n  workflow_dispatch:\n    inputs:\n      v:\n        type: string\njobs: {}\n"
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "GET", path: "/repos/o/r/actions/workflows", body: wfList},
		route{method: "GET", path: "/repos/o/r/contents/.github/workflows/ci.yml", body: contentJSON(yml)}))
	dw, err := c.ListDispatchableWorkflows(context.Background(), "o", "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(dw) != 1 || dw[0].Name != "CI" || len(dw[0].Inputs) != 1 {
		t.Fatalf("dispatchable wrong: %+v", dw)
	}
}

func TestRunJobGraph(t *testing.T) {
	yml := "jobs:\n  build:\n    runs-on: x\n  test:\n    needs: build\n"
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "GET", path: "/repos/o/r/actions/runs/100", body: `{"id":100,"workflow_id":1}`},
		route{method: "GET", path: "/repos/o/r/actions/workflows/1", body: `{"id":1,"path":".github/workflows/ci.yml"}`},
		route{method: "GET", path: "/repos/o/r/contents/.github/workflows/ci.yml", body: contentJSON(yml)}))
	nodes, err := c.RunJobGraph(context.Background(), "o", "r", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 job nodes, got %d: %+v", len(nodes), nodes)
	}
}
