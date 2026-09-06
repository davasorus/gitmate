package ghapi

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
	"time"
)

func TestListWorkflows(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/actions/workflows",
		`{"total_count":1,"workflows":[{"id":1,"name":"CI","state":"active","path":".github/workflows/ci.yml"}]}`))
	wfs, err := c.ListWorkflows(context.Background(), "o", "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(wfs) != 1 || wfs[0].Name != "CI" {
		t.Fatalf("workflows wrong: %+v", wfs)
	}
}

func TestListRuns(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/actions/runs",
		`{"total_count":1,"workflow_runs":[{"id":100,"name":"CI","status":"completed","conclusion":"success","head_branch":"main","event":"push","run_number":5,"workflow_id":1,"html_url":"http://x/100"}]}`))
	runs, err := c.ListRuns(context.Background(), "o", "r", 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].ID != 100 || runs[0].Conclusion != "success" {
		t.Fatalf("runs wrong: %+v", runs)
	}
}

func TestGetRun(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/actions/runs/100",
		`{"id":100,"name":"CI","status":"in_progress","head_branch":"main","run_number":5}`))
	run, err := c.GetRun(context.Background(), "o", "r", 100)
	if err != nil {
		t.Fatal(err)
	}
	if run.ID != 100 || run.Status != "in_progress" {
		t.Fatalf("run wrong: %+v", run)
	}
}

func TestListRunJobs(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/actions/runs/100/jobs",
		`{"total_count":1,"jobs":[{"id":200,"name":"build","status":"completed","conclusion":"success","steps":[{"name":"checkout","status":"completed","conclusion":"success","number":1}]}]}`))
	jobs, err := c.ListRunJobs(context.Background(), "o", "r", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].Name != "build" || len(jobs[0].Steps) != 1 {
		t.Fatalf("jobs wrong: %+v", jobs)
	}
}

func TestCancelRun(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t, route{method: "POST", path: "/repos/o/r/actions/runs/100/cancel", status: 202}))
	if err := c.CancelRun(context.Background(), "o", "r", 100); err != nil {
		t.Fatal(err)
	}
}

func TestRerunRun(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t, route{method: "POST", path: "/repos/o/r/actions/runs/100/rerun", status: 201}))
	if err := c.RerunRun(context.Background(), "o", "r", 100); err != nil {
		t.Fatal(err)
	}
}

func TestRerunFailed(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t, route{method: "POST", path: "/repos/o/r/actions/runs/100/rerun-failed-jobs", status: 201}))
	if err := c.RerunFailed(context.Background(), "o", "r", 100); err != nil {
		t.Fatal(err)
	}
}

func TestTriggerDispatch(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t, route{method: "POST", path: "/repos/o/r/actions/workflows/ci.yml/dispatches", status: 204}))
	if err := c.TriggerDispatch(context.Background(), "o", "r", "ci.yml", "main", map[string]interface{}{"k": "v"}); err != nil {
		t.Fatal(err)
	}
}

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

func TestJobLogs(t *testing.T) {
	logText := "##[group]Setup\nprep\n##[endgroup]\n##[group]Run\nbuilding\n##[endgroup]\n"
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/o/r/actions/jobs/9/logs":
			http.Redirect(w, r, "http://"+r.Host+"/logblob", http.StatusFound)
		case "/logblob":
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
}

// --- pure parser/helper tests (parseDispatchInputs, splitStepLogs, parseJobGraph, humanDur) ---

func TestParseDispatchInputs(t *testing.T) {
	yml := `
on:
  workflow_dispatch:
    inputs:
      version:
        description: "the version"
        required: true
        default: "1.0"
        type: string
      env:
        type: choice
        options: [dev, prod]
jobs:
  build:
    runs-on: ubuntu-latest
`
	inputs, has := parseDispatchInputs(yml)
	if !has {
		t.Fatal("expected hasDispatch true")
	}
	if len(inputs) != 2 || inputs[0].Name != "version" || !inputs[0].Required || inputs[0].Default != "1.0" {
		t.Fatalf("version input wrong: %+v", inputs)
	}
	if inputs[1].Type != "choice" || len(inputs[1].Options) != 2 {
		t.Errorf("env choice wrong: %+v", inputs[1])
	}
}

func TestParseDispatchInputs_NoDispatch(t *testing.T) {
	if _, has := parseDispatchInputs("on:\n  push:\njobs: {}\n"); has {
		t.Fatal("expected hasDispatch false for push-only")
	}
}

func TestParseDispatchInputs_ScalarOn(t *testing.T) {
	if _, has := parseDispatchInputs("on: workflow_dispatch\njobs: {}\n"); !has {
		t.Fatal("expected hasDispatch true for scalar on: workflow_dispatch")
	}
}

func TestSplitStepLogs(t *testing.T) {
	raw := "##[group]Run checkout\nchecked out\n##[endgroup]\n##[group]Build\ncompiling\ndone\n##[endgroup]\n"
	steps := splitStepLogs(raw)
	if len(steps) != 2 || steps[0].Name != "Run checkout" || steps[1].Name != "Build" {
		t.Fatalf("split wrong: %+v", steps)
	}
}

func TestParseJobGraph(t *testing.T) {
	yml := "jobs:\n  build:\n    runs-on: x\n  test:\n    needs: build\n  deploy:\n    needs: [build, test]\n"
	nodes := parseJobGraph(yml)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(nodes))
	}
	byName := map[string][]string{}
	for _, n := range nodes {
		byName[n.Name] = n.Needs
	}
	if len(byName["build"]) != 0 || len(byName["test"]) != 1 || len(byName["deploy"]) != 2 {
		t.Errorf("needs wrong: %+v", byName)
	}
}

func TestHumanDur(t *testing.T) {
	cases := map[int]string{5: "5s", 65: "1m5s", 3661: "61m1s"}
	for secs, want := range cases {
		if got := humanDur(time.Duration(secs) * time.Second); got != want {
			t.Errorf("humanDur(%ds) = %q, want %q", secs, got, want)
		}
	}
}

func TestListRunsWithDuration(t *testing.T) {
	// a completed run with run_started_at + updated_at exercises toRun's duration calc
	body := `{"total_count":1,"workflow_runs":[{
		"id":101,"name":"CI","status":"completed","conclusion":"success",
		"head_branch":"main","event":"push","run_number":6,"workflow_id":1,
		"run_started_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:01:23Z",
		"created_at":"2026-01-01T00:00:00Z","html_url":"http://x/101"}]}`
	c, _ := newTestClient(t, jsonHandler(t, "/repos/o/r/actions/runs", body))
	runs, err := c.ListRuns(context.Background(), "o", "r", 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].Duration == "" {
		t.Fatalf("expected a computed duration, got %+v", runs[0])
	}
}
