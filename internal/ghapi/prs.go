package ghapi

import (
	"context"
	"time"

	"github.com/google/go-github/v66/github"
)

// PR is a trimmed pull-request view.
type PR struct {
	Number int
	Title  string
	Author string
	State  string
	When   time.Time
	Draft  bool
}

// Issue is a trimmed issue view.
type Issue struct {
	Number int
	Title  string
	Author string
	State  string
	When   time.Time
}

// ListPRs returns open pull requests for owner/repo.
func (c *Client) ListPRs(ctx context.Context, owner, repo, state string) ([]PR, error) {
	opts := &github.PullRequestListOptions{
		State:       state, // "open", "closed", or "all"
		ListOptions: github.ListOptions{PerPage: 30},
	}
	raw, _, err := c.gh.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	var prs []PR
	for _, p := range raw {
		prs = append(prs, PR{
			Number: p.GetNumber(),
			Title:  p.GetTitle(),
			Author: p.GetUser().GetLogin(),
			State:  p.GetState(),
			When:   p.GetCreatedAt().Time,
			Draft:  p.GetDraft(),
		})
	}
	return prs, nil
}

// ListIssues returns issues for owner/repo — with PRs filtered out.
func (c *Client) ListIssues(ctx context.Context, owner, repo, state string) ([]Issue, error) {
	opts := &github.IssueListByRepoOptions{
		State:       state,
		ListOptions: github.ListOptions{PerPage: 30},
	}
	raw, _, err := c.gh.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	var issues []Issue
	for _, i := range raw {
		// GitHub's Issues API returns PRs too. Skip anything that is a PR.
		if i.IsPullRequest() {
			continue
		}
		issues = append(issues, Issue{
			Number: i.GetNumber(),
			Title:  i.GetTitle(),
			Author: i.GetUser().GetLogin(),
			State:  i.GetState(),
			When:   i.GetCreatedAt().Time,
		})
	}
	return issues, nil
}

func labelNames(ls []*github.Label) []string {
	names := make([]string, 0, len(ls))
	for _, l := range ls {
		names = append(names, l.GetName())
	}
	return names
}
