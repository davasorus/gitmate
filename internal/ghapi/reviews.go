package ghapi

import (
	"context"

	"github.com/google/go-github/v88/github"
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
	rr, _, err := c.gh.PullRequests.ListReviewers(ctx, owner, repo, number)
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

// ExistingComment is a review comment already posted on the PR, anchored to a
// file + line. ReplyToID (0 if top-level) links a reply to its parent thread.
type ExistingComment struct {
	ID        int64
	Path      string
	Line      int
	Author    string
	Body      string
	InReplyTo int64
}

// IssueComment is a general (not line-anchored) PR comment from the issue stream.
type IssueComment struct {
	ID     int64
	Author string
	Body   string
}

// ListReviewComments returns the line-anchored review comments on a PR.
func (c *Client) ListReviewComments(ctx context.Context, owner, repo string, number int) ([]ExistingComment, error) {
	var out []ExistingComment
	opts := &github.PullRequestListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		raw, resp, err := c.gh.PullRequests.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		for _, cm := range raw {
			out = append(out, ExistingComment{
				ID:        cm.GetID(),
				Path:      cm.GetPath(),
				Line:      cm.GetLine(),
				Author:    cm.GetUser().GetLogin(),
				Body:      cm.GetBody(),
				InReplyTo: cm.GetInReplyTo(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// ListIssueComments returns the general (non-line) comment stream on a PR.
func (c *Client) ListIssueComments(ctx context.Context, owner, repo string, number int) ([]IssueComment, error) {
	var out []IssueComment
	opts := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		raw, resp, err := c.gh.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		for _, cm := range raw {
			out = append(out, IssueComment{ID: cm.GetID(), Author: cm.GetUser().GetLogin(), Body: cm.GetBody()})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// ReplyToReviewComment posts a reply to an existing review comment thread.
func (c *Client) ReplyToReviewComment(ctx context.Context, owner, repo string, number int, commentID int64, body string) error {
	_, _, err := c.gh.PullRequests.CreateCommentInReplyTo(ctx, owner, repo, number, body, commentID)
	return err
}
