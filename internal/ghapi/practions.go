package ghapi

import (
	"context"

	"github.com/google/go-github/v75/github"
)

// CheckRun is a trimmed view of one CI check on a commit.
type CheckRun struct {
	Name       string
	Status     string // queued, in_progress, completed
	Conclusion string // success, failure, neutral, cancelled, ... (empty until completed)
}

// MergePR merges a pull request. method is "merge", "squash", or "rebase".
// Returns the merge commit SHA on success.
func (c *Client) MergePR(ctx context.Context, owner, repo string, number int, method string) (string, error) {
	if method == "" {
		method = "merge"
	}
	res, _, err := c.gh.PullRequests.Merge(ctx, owner, repo, number, "", &github.PullRequestOptions{
		MergeMethod: method,
	})
	if err != nil {
		return "", err
	}
	return res.GetSHA(), nil
}

// CommentPR posts a comment on a PR or issue (they share the comment endpoint).
func (c *Client) CommentPR(ctx context.Context, owner, repo string, number int, body string) (string, error) {
	comment, _, err := c.gh.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: github.String(body),
	})
	if err != nil {
		return "", err
	}
	return comment.GetHTMLURL(), nil
}

// PRChecks returns the CI check runs for a PR's head commit — this is how you
// see whether GitHub Actions (your own CI) passed on the PR.
func (c *Client) PRChecks(ctx context.Context, owner, repo string, number int) ([]CheckRun, error) {
	// A check run is attached to a commit, so first resolve the PR's head SHA.
	pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	ref := pr.GetHead().GetSHA()

	results, _, err := c.gh.Checks.ListCheckRunsForRef(ctx, owner, repo, ref, &github.ListCheckRunsOptions{})
	if err != nil {
		return nil, err
	}
	var runs []CheckRun
	for _, r := range results.CheckRuns {
		runs = append(runs, CheckRun{
			Name:       r.GetName(),
			Status:     r.GetStatus(),
			Conclusion: r.GetConclusion(),
		})
	}
	return runs, nil
}
