package ghapi

import (
	"context"
	"testing"
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
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/actions/runs/100/cancel", status: 202}))
	if err := c.CancelRun(context.Background(), "o", "r", 100); err != nil {
		t.Fatal(err)
	}
}

func TestRerunRun(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/actions/runs/100/rerun", status: 201}))
	if err := c.RerunRun(context.Background(), "o", "r", 100); err != nil {
		t.Fatal(err)
	}
}

func TestRerunFailed(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/actions/runs/100/rerun-failed-jobs", status: 201}))
	if err := c.RerunFailed(context.Background(), "o", "r", 100); err != nil {
		t.Fatal(err)
	}
}

func TestTriggerDispatch(t *testing.T) {
	c, _ := newTestClient(t, routeHandler(t,
		route{method: "POST", path: "/repos/o/r/actions/workflows/ci.yml/dispatches", status: 204}))
	err := c.TriggerDispatch(context.Background(), "o", "r", "ci.yml", "main", map[string]interface{}{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
}
