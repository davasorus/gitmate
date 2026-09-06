package ghapi

import (
	"context"

	"github.com/google/go-github/v66/github"
)

// CreatePR opens a pull request from head into base. head is a branch name
// (e.g. "feature/x"); base is usually the default branch (e.g. "live").
// Returns the new PR number and its HTML URL.
func (c *Client) CreatePR(ctx context.Context, owner, repo, title, body, head, base string) (int, string, error) {
	pr, _, err := c.gh.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	})
	if err != nil {
		return 0, "", err
	}
	return pr.GetNumber(), pr.GetHTMLURL(), nil
}

// CreateIssue opens an issue. Returns the new issue number and its HTML URL.
func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string) (int, string, error) {
	issue, _, err := c.gh.Issues.Create(ctx, owner, repo, &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	})
	if err != nil {
		return 0, "", err
	}
	return issue.GetNumber(), issue.GetHTMLURL(), nil
}

// SetIssueState closes or reopens an issue. state is "closed" or "open".
func (c *Client) SetIssueState(ctx context.Context, owner, repo string, number int, state string) error {
	_, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, &github.IssueRequest{
		State: github.String(state),
	})
	return err
}

// SetPRState closes or reopens a pull request. state is "closed" or "open".
func (c *Client) SetPRState(ctx context.Context, owner, repo string, number int, state string) error {
	_, _, err := c.gh.PullRequests.Edit(ctx, owner, repo, number, &github.PullRequest{
		State: github.String(state),
	})
	return err
}
