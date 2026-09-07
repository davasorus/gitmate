package ghapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v88/github"
	"gopkg.in/yaml.v3"
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

// CancelRun cancels an in-progress workflow run.
func (c *Client) CancelRun(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.gh.Actions.CancelWorkflowRunByID(ctx, owner, repo, runID)
	// GitHub returns 202 Accepted for cancel; go-github surfaces that as
	// *AcceptedError. For an async cancel, "accepted" IS success.
	if _, ok := err.(*github.AcceptedError); ok {
		return nil
	}
	return err
}

// RerunRun re-runs all jobs of a completed run.
func (c *Client) RerunRun(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.gh.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	return err
}

// RerunFailed re-runs only the failed jobs of a completed run.
func (c *Client) RerunFailed(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.gh.Actions.RerunFailedJobsByID(ctx, owner, repo, runID)
	return err
}

// DispatchInput describes one workflow_dispatch input the workflow defines.
type DispatchInput struct {
	Name        string
	Description string
	Required    bool
	Default     string
	Type        string   // string, boolean, choice, number, environment
	Options     []string // for type=choice
}

// DispatchableWorkflow is a workflow that has a workflow_dispatch trigger, with
// its declared inputs (parsed from the workflow YAML).
type DispatchableWorkflow struct {
	ID     int64
	Name   string
	Path   string
	Inputs []DispatchInput
}

// TriggerDispatch fires a workflow_dispatch event on the given workflow file
// (e.g. "release.yml") for a ref (branch/tag), with input values.
func (c *Client) TriggerDispatch(ctx context.Context, owner, repo, workflowFile, ref string, inputs map[string]interface{}) error {
	_, _, err := c.gh.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowFile,
		github.CreateWorkflowDispatchEventRequest{Ref: ref, Inputs: inputs})
	return err
}

// ListDispatchableWorkflows returns workflows that declare a workflow_dispatch
// trigger, with their inputs parsed from the workflow YAML. This drives a dynamic
// dispatch form (render a field per declared input).
func (c *Client) ListDispatchableWorkflows(ctx context.Context, owner, repo string) ([]DispatchableWorkflow, error) {
	wf, _, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}
	var out []DispatchableWorkflow
	for _, w := range wf.Workflows {
		content, _, _, cerr := c.gh.Repositories.GetContents(ctx, owner, repo, w.GetPath(), nil)
		if cerr != nil || content == nil {
			continue
		}
		raw, derr := content.GetContent()
		if derr != nil {
			continue
		}
		inputs, hasDispatch := parseDispatchInputs(raw)
		if !hasDispatch {
			continue
		}
		out = append(out, DispatchableWorkflow{ID: w.GetID(), Name: w.GetName(), Path: w.GetPath(), Inputs: inputs})
	}
	return out, nil
}

// parseDispatchInputs reads a workflow YAML and, if it has an on.workflow_dispatch
// trigger, returns its declared inputs. hasDispatch is false if the workflow has
// no workflow_dispatch trigger at all.
func parseDispatchInputs(yml string) (inputs []DispatchInput, hasDispatch bool) {
	var doc struct {
		// "on" is a YAML keyword-ish; capture it as a generic node.
		On yaml.Node `yaml:"on"`
	}
	if err := yaml.Unmarshal([]byte(yml), &doc); err != nil {
		return nil, false
	}
	// on: can be a string ("push"), a list, or a map. We want the map form with
	// a workflow_dispatch key that may carry inputs.
	if doc.On.Kind != yaml.MappingNode {
		// could still be "on: workflow_dispatch" (scalar) or a sequence containing it
		if doc.On.Kind == yaml.ScalarNode && doc.On.Value == "workflow_dispatch" {
			return nil, true
		}
		if doc.On.Kind == yaml.SequenceNode {
			for _, n := range doc.On.Content {
				if n.Value == "workflow_dispatch" {
					return nil, true
				}
			}
		}
		return nil, false
	}
	for i := 0; i+1 < len(doc.On.Content); i += 2 {
		key := doc.On.Content[i]
		val := doc.On.Content[i+1]
		if key.Value != "workflow_dispatch" {
			continue
		}
		// found the workflow_dispatch trigger; find its inputs
		if val.Kind != yaml.MappingNode {
			return nil, true
		}
		for j := 0; j+1 < len(val.Content); j += 2 {
			if val.Content[j].Value != "inputs" {
				continue
			}
			inNode := val.Content[j+1]
			if inNode.Kind != yaml.MappingNode {
				break
			}
			for k := 0; k+1 < len(inNode.Content); k += 2 {
				name := inNode.Content[k].Value
				spec := inNode.Content[k+1]
				di := DispatchInput{Name: name, Type: "string"}
				if spec.Kind == yaml.MappingNode {
					for m := 0; m+1 < len(spec.Content); m += 2 {
						sk := spec.Content[m].Value
						sv := spec.Content[m+1]
						switch sk {
						case "description":
							di.Description = sv.Value
						case "required":
							di.Required = sv.Value == "true"
						case "default":
							di.Default = sv.Value
						case "type":
							di.Type = sv.Value
						case "options":
							for _, o := range sv.Content {
								di.Options = append(di.Options, o.Value)
							}
						}
					}
				}
				inputs = append(inputs, di)
			}
			break
		}
		return inputs, true
	}
	return inputs, hasDispatch
}

// JobLog is a job's captured log, split (best-effort) into per-step sections.
type JobLog struct {
	JobName string
	Steps   []StepLog
	Raw     string // full log; shown when per-step split isn't reliable
}

type StepLog struct {
	Name string
	Text string
}

// JobLogs downloads a single job's log (completed runs only) and returns it,
// best-effort split per step using GitHub's "##[group]" markers. On any parse
// uncertainty the Raw whole-job log is always available as a fallback.
func (c *Client) JobLogs(ctx context.Context, owner, repo string, jobID int64) (JobLog, error) {
	u, _, err := c.gh.Actions.GetWorkflowJobLogs(ctx, owner, repo, jobID, 3)
	if err != nil {
		return JobLog{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return JobLog{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return JobLog{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return JobLog{}, err
	}
	raw := string(data)
	return JobLog{Steps: splitStepLogs(raw), Raw: raw}, nil
}

// splitStepLogs splits a job log into steps using the "##[group]" / "##[endgroup]"
// markers GitHub emits around each step. Falls back to a single section if none.
func splitStepLogs(raw string) []StepLog {
	lines := strings.Split(raw, "\n")
	var steps []StepLog
	var cur *StepLog
	for _, ln := range lines {
		// timestamps prefix each line: "2026-...Z ##[group]Run actions/checkout"
		marker := ln
		if i := strings.Index(ln, "##[group]"); i >= 0 {
			name := strings.TrimSpace(ln[i+len("##[group]"):])
			if cur != nil {
				steps = append(steps, *cur)
			}
			cur = &StepLog{Name: name}
			continue
		}
		if strings.Contains(marker, "##[endgroup]") {
			continue
		}
		if cur != nil {
			cur.Text += ln + "\n"
		}
	}
	if cur != nil {
		steps = append(steps, *cur)
	}
	return steps
}

// JobNode is a job in the workflow's dependency graph: its name and the jobs it
// needs (must complete first). Parsed from the workflow YAML.
type JobNode struct {
	Name  string
	Needs []string
}

// RunJobGraph returns the job dependency graph for a run's workflow, parsed from
// the workflow YAML (jobs.<name>.needs). The GUI overlays live job status onto
// this to draw a flowchart. Falls back to a flat (no-edges) graph if YAML parse
// fails or needs aren't declared.
func (c *Client) RunJobGraph(ctx context.Context, owner, repo string, runID int64) ([]JobNode, error) {
	rn, _, err := c.gh.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return nil, err
	}
	wfID := rn.GetWorkflowID()
	wf, _, err := c.gh.Actions.GetWorkflowByID(ctx, owner, repo, wfID)
	if err != nil {
		return nil, err
	}
	content, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, wf.GetPath(), nil)
	if err != nil || content == nil {
		return nil, err
	}
	raw, err := content.GetContent()
	if err != nil {
		return nil, err
	}
	return parseJobGraph(raw), nil
}

// parseJobGraph reads a workflow YAML and returns its jobs with their needs.
func parseJobGraph(yml string) []JobNode {
	var doc struct {
		Jobs yaml.Node `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(yml), &doc); err != nil {
		return nil
	}
	if doc.Jobs.Kind != yaml.MappingNode {
		return nil
	}

	// First pass: read each declared job's needs + matrix legs (base names).
	type rawJob struct {
		needs []string
		legs  []string // matrix leg labels, e.g. "ubuntu-latest, 1.25"; empty if no matrix
	}
	raw := map[string]*rawJob{}
	var order []string
	for i := 0; i+1 < len(doc.Jobs.Content); i += 2 {
		name := doc.Jobs.Content[i].Value
		spec := doc.Jobs.Content[i+1]
		rj := &rawJob{}
		if spec.Kind == yaml.MappingNode {
			for j := 0; j+1 < len(spec.Content); j += 2 {
				key := spec.Content[j].Value
				val := spec.Content[j+1]
				switch key {
				case "needs":
					switch val.Kind {
					case yaml.ScalarNode:
						rj.needs = append(rj.needs, val.Value)
					case yaml.SequenceNode:
						for _, n := range val.Content {
							rj.needs = append(rj.needs, n.Value)
						}
					}
				case "strategy":
					rj.legs = matrixLegs(val)
				}
			}
		}
		raw[name] = rj
		order = append(order, name)
	}

	// expandedNames: for a base job, the runtime node name(s) it becomes.
	// GitHub names matrix runtime jobs "base (leg1, leg2, ...)".
	expanded := func(base string) []string {
		rj := raw[base]
		if rj == nil || len(rj.legs) == 0 {
			return []string{base}
		}
		names := make([]string, 0, len(rj.legs))
		for _, leg := range rj.legs {
			names = append(names, base+" ("+leg+")")
		}
		return names
	}

	// Second pass: emit one node per (expanded) runtime job; fan needs edges to
	// every expanded leg of each dependency.
	var out []JobNode
	for _, base := range order {
		rj := raw[base]
		var deps []string
		for _, dep := range rj.needs {
			deps = append(deps, expanded(dep)...)
		}
		for _, nodeName := range expanded(base) {
			out = append(out, JobNode{Name: nodeName, Needs: deps})
		}
	}
	return out
}

// matrixLegs extracts matrix combinations from a job's `strategy:` node and
// returns one label per combination (the parenthesized text GitHub appends to
// the runtime job name, e.g. "ubuntu-latest, 1.25"). Returns nil if no matrix.
// matrixCombo is an ordered set of axis key→value pairs (order preserved so the
// generated label matches GitHub's "base (v1, v2, ...)" runtime job naming).
type matrixCombo struct {
	keys []string
	vals map[string]string
}

func (c matrixCombo) clone() matrixCombo {
	nc := matrixCombo{keys: append([]string(nil), c.keys...), vals: map[string]string{}}
	for k, v := range c.vals {
		nc.vals[k] = v
	}
	return nc
}

// label renders the parenthesized text GitHub appends to a matrix job's name,
// e.g. "ubuntu-latest, 1.25" — axis values in declaration order.
func (c matrixCombo) label() string {
	parts := make([]string, 0, len(c.keys))
	for _, k := range c.keys {
		parts = append(parts, c.vals[k])
	}
	return strings.Join(parts, ", ")
}

// matrixLegs computes the matrix combinations for a job's strategy, faithfully
// applying GitHub's rules: cartesian product of axes, minus `exclude` matches,
// plus `include` (which extends a matching combo or appends a new one). Returns
// one label per resulting runtime job, or nil if there's no matrix.
func matrixLegs(strategy *yaml.Node) []string {
	if strategy == nil || strategy.Kind != yaml.MappingNode {
		return nil
	}
	var matrix *yaml.Node
	for i := 0; i+1 < len(strategy.Content); i += 2 {
		if strategy.Content[i].Value == "matrix" {
			matrix = strategy.Content[i+1]
		}
	}
	if matrix == nil || matrix.Kind != yaml.MappingNode {
		return nil
	}

	// 1) gather axes (in declaration order) + include/exclude entries
	type axis struct {
		key    string
		values []string
	}
	var axes []axis
	var includes, excludes []map[string]string
	for i := 0; i+1 < len(matrix.Content); i += 2 {
		key := matrix.Content[i].Value
		val := matrix.Content[i+1]
		switch key {
		case "include":
			includes = mapEntriesFromSeq(val)
		case "exclude":
			excludes = mapEntriesFromSeq(val)
		default:
			if val.Kind == yaml.SequenceNode {
				a := axis{key: key}
				for _, v := range val.Content {
					a.values = append(a.values, v.Value)
				}
				if len(a.values) > 0 {
					axes = append(axes, a)
				}
			}
		}
	}

	// 2) cartesian product of the axes → base combos
	combos := []matrixCombo{{vals: map[string]string{}}}
	for _, a := range axes {
		var next []matrixCombo
		for _, c := range combos {
			for _, v := range a.values {
				nc := c.clone()
				nc.keys = append(nc.keys, a.key)
				nc.vals[a.key] = v
				next = append(next, nc)
			}
		}
		combos = next
	}

	// 3) apply exclude: drop combos that match ALL key/values of any exclude entry
	if len(excludes) > 0 {
		var kept []matrixCombo
		for _, c := range combos {
			excluded := false
			for _, ex := range excludes {
				match := true
				for k, v := range ex {
					if c.vals[k] != v {
						match = false
						break
					}
				}
				if match && len(ex) > 0 {
					excluded = true
					break
				}
			}
			if !excluded {
				kept = append(kept, c)
			}
		}
		combos = kept
	}

	// 4) apply include: for each include entry, if its keys that overlap existing
	// axes match some combo(s), extend those with the new keys; otherwise append
	// it as a standalone combo. (Simplified from GitHub's exact algorithm, but
	// covers the common cases: adding a field to a combo, or adding a new combo.)
	for _, inc := range includes {
		if len(inc) == 0 {
			continue
		}
		extendedAny := false
		for ci := range combos {
			c := &combos[ci]
			// does this include entry's overlapping keys match this combo?
			overlaps := false
			match := true
			for k, v := range inc {
				if _, isAxis := c.vals[k]; isAxis {
					overlaps = true
					if c.vals[k] != v {
						match = false
						break
					}
				}
			}
			if overlaps && match {
				extendedAny = true
				for k, v := range inc {
					if _, exists := c.vals[k]; !exists {
						c.keys = append(c.keys, k)
						c.vals[k] = v
					}
				}
			}
		}
		if !extendedAny {
			nc := matrixCombo{vals: map[string]string{}}
			for k, v := range inc {
				nc.keys = append(nc.keys, k)
				nc.vals[k] = v
			}
			combos = append(combos, nc)
		}
	}

	if len(combos) == 0 {
		return nil
	}
	out := make([]string, 0, len(combos))
	for _, c := range combos {
		if lbl := c.label(); lbl != "" {
			out = append(out, lbl)
		}
	}
	return out
}

// mapEntriesFromSeq reads a YAML sequence of mappings (include:/exclude:) into
// a slice of key→value maps.
func mapEntriesFromSeq(seq *yaml.Node) []map[string]string {
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil
	}
	var out []map[string]string
	for _, item := range seq.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}
		m := map[string]string{}
		for i := 0; i+1 < len(item.Content); i += 2 {
			m[item.Content[i].Value] = item.Content[i+1].Value
		}
		if len(m) > 0 {
			out = append(out, m)
		}
	}
	return out
}
