package ghapi

import (
	"context"

	"github.com/google/go-github/v88/github"
)

// Milestone is a repository milestone (number + title + open/closed state).
type Milestone struct {
	Number int
	Title  string
	State  string
}

// ListMilestones returns the repository's milestones (open by default; pass
// "all" or "closed" for others).
func (c *Client) ListMilestones(ctx context.Context, owner, repo, state string) ([]Milestone, error) {
	if state == "" {
		state = "open"
	}
	raw, _, err := c.gh.Issues.ListMilestones(ctx, owner, repo,
		&github.MilestoneListOptions{State: state, ListOptions: github.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, err
	}
	var out []Milestone
	for _, m := range raw {
		out = append(out, Milestone{Number: m.GetNumber(), Title: m.GetTitle(), State: m.GetState()})
	}
	return out, nil
}

// SetMilestone assigns an issue/PR to a milestone by number; pass 0 to clear it.
func (c *Client) SetMilestone(ctx context.Context, owner, repo string, number, milestone int) error {
	req := &github.IssueRequest{}
	if milestone == 0 {
		// clearing requires an explicit null milestone via the raw field
		req.Milestone = nil
	} else {
		req.Milestone = github.Ptr(milestone)
	}
	_, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, req)
	return err
}
