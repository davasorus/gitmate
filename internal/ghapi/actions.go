package ghapi

import (
	"context"
	"time"

	"github.com/google/go-github/v66/github"
)

// Workflow is a repo workflow definition.
type Workflow struct {
	ID    int64
	Name  string
	State string
	Path  string
}

// WorkflowRun is one execution of a workflow.
type WorkflowRun struct {
	ID           int64
	Name         string
	WorkflowID   int64
	WorkflowName string // for grouping (falls back to Name)
	Status       string // queued, in_progress, completed
	Conclusion   string // success, failure, cancelled, "" while running
	Branch       string
	Event        string
	Number       int
	CreatedAt    string
	Duration     string // human-readable run duration ("1m23s"), "" if not finished/started
	URL          string
}

// Job is a job within a run; Steps are its steps.
type Job struct {
	ID         int64
	Name       string
	Status     string
	Conclusion string
	Steps      []Step
}

type Step struct {
	Name       string
	Status     string
	Conclusion string
	Number     int
}

// ListWorkflows returns the repo's workflow definitions.
func (c *Client) ListWorkflows(ctx context.Context, owner, repo string) ([]Workflow, error) {
	wf, _, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}
	var out []Workflow
	for _, w := range wf.Workflows {
		out = append(out, Workflow{ID: w.GetID(), Name: w.GetName(), State: w.GetState(), Path: w.GetPath()})
	}
	return out, nil
}

// ListRuns returns recent workflow runs for the repo, newest first.
func (c *Client) ListRuns(ctx context.Context, owner, repo string, limit int) ([]WorkflowRun, error) {
	if limit <= 0 {
		limit = 30
	}
	runs, _, err := c.gh.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo,
		&github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: limit}})
	if err != nil {
		return nil, err
	}
	var out []WorkflowRun
	for _, r := range runs.WorkflowRuns {
		out = append(out, toRun(r))
	}
	return out, nil
}

// GetRun returns a single run's summary.
func (c *Client) GetRun(ctx context.Context, owner, repo string, runID int64) (WorkflowRun, error) {
	r, _, err := c.gh.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return WorkflowRun{}, err
	}
	return toRun(r), nil
}

// ListRunJobs returns the jobs (with steps) for a run.
func (c *Client) ListRunJobs(ctx context.Context, owner, repo string, runID int64) ([]Job, error) {
	jobs, _, err := c.gh.Actions.ListWorkflowJobs(ctx, owner, repo, runID,
		&github.ListWorkflowJobsOptions{ListOptions: github.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, err
	}
	var out []Job
	for _, j := range jobs.Jobs {
		job := Job{ID: j.GetID(), Name: j.GetName(), Status: j.GetStatus(), Conclusion: j.GetConclusion()}
		for _, st := range j.Steps {
			job.Steps = append(job.Steps, Step{
				Name: st.GetName(), Status: st.GetStatus(), Conclusion: st.GetConclusion(), Number: int(st.GetNumber()),
			})
		}
		out = append(out, job)
	}
	return out, nil
}

func toRun(r *github.WorkflowRun) WorkflowRun {
	name := r.GetName()
	dur := ""
	if !r.GetRunStartedAt().IsZero() && r.GetStatus() == "completed" {
		d := r.GetUpdatedAt().Sub(r.GetRunStartedAt().Time)
		if d > 0 {
			dur = humanDur(d)
		}
	}
	return WorkflowRun{
		ID:           r.GetID(),
		Name:         name,
		WorkflowID:   r.GetWorkflowID(),
		WorkflowName: name, // run name == workflow name for most triggers; grouped by WorkflowID
		Status:       r.GetStatus(),
		Conclusion:   r.GetConclusion(),
		Branch:       r.GetHeadBranch(),
		Event:        r.GetEvent(),
		Number:       r.GetRunNumber(),
		CreatedAt:    r.GetCreatedAt().Format("2006-01-02 15:04"),
		Duration:     dur,
		URL:          r.GetHTMLURL(),
	}
}

func humanDur(d time.Duration) string {
	s := int(d.Seconds())
	if s < 60 {
		return fmtInt(s) + "s"
	}
	m := s / 60
	s = s % 60
	return fmtInt(m) + "m" + fmtInt(s) + "s"
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
