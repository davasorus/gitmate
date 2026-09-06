package ghapi

import (
	"context"

	"github.com/google/go-github/v66/github"
)

// Review is a trimmed PR review (a whole-PR verdict).
type Review struct {
	ID     int64
	Author string
	State  string // APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED, PENDING
	Body   string
}

// ReviewComment is a pending line comment to attach to a submitted review.
// Path is the file path; Line is the line number in the file's new version;
// Body is the comment text.
type ReviewComment struct {
	Path string
	Line int
	Body string
}

// Reviewer is a requested (not-yet-submitted) reviewer on a PR.
type Reviewer struct {
	Login string
}

// ListReviews returns the submitted reviews on a PR, oldest first.
func (c *Client) ListReviews(ctx context.Context, owner, repo string, number int) ([]Review, error) {
	var out []Review
	opts := &github.ListOptions{PerPage: 50}
	for {
		raw, resp, err := c.gh.PullRequests.ListReviews(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		for _, r := range raw {
			out = append(out, Review{
				ID:     r.GetID(),
				Author: r.GetUser().GetLogin(),
				State:  r.GetState(),
				Body:   r.GetBody(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// SubmitReview posts a whole-PR review. event is "APPROVE",
// "REQUEST_CHANGES", or "COMMENT". body is optional for APPROVE but required
// by GitHub for REQUEST_CHANGES and COMMENT.
func (c *Client) SubmitReview(ctx context.Context, owner, repo string, number int, event, body string, comments []ReviewComment) error {
	req := &github.PullRequestReviewRequest{
		Event: github.String(event),
		Body:  github.String(body),
	}
	for _, cm := range comments {
		line := cm.Line
		req.Comments = append(req.Comments, &github.DraftReviewComment{
			Path: github.String(cm.Path),
			Line: github.Int(line),
			Body: github.String(cm.Body),
			Side: github.String("RIGHT"),
		})
	}
	_, _, err := c.gh.PullRequests.CreateReview(ctx, owner, repo, number, req)
	return err
}

// ListRequestedReviewers returns users whose review has been requested but not
// yet submitted.
func (c *Client) ListRequestedReviewers(ctx context.Context, owner, repo string, number int) ([]Reviewer, error) {
	rr, _, err := c.gh.PullRequests.ListReviewers(ctx, owner, repo, number, &github.ListOptions{PerPage: 50})
	if err != nil {
		return nil, err
	}
	var out []Reviewer
	for _, u := range rr.Users {
		out = append(out, Reviewer{Login: u.GetLogin()})
	}
	return out, nil
}

// RequestReviewers asks the given users to review the PR.
func (c *Client) RequestReviewers(ctx context.Context, owner, repo string, number int, logins []string) error {
	_, _, err := c.gh.PullRequests.RequestReviewers(ctx, owner, repo, number,
		github.ReviewersRequest{Reviewers: logins})
	return err
}

// RemoveReviewer withdraws a review request from a user.
func (c *Client) RemoveReviewer(ctx context.Context, owner, repo string, number int, login string) error {
	_, err := c.gh.PullRequests.RemoveReviewers(ctx, owner, repo, number,
		github.ReviewersRequest{Reviewers: []string{login}})
	return err
}
