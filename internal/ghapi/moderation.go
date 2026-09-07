package ghapi

import (
	"context"

	"github.com/google/go-github/v88/github"
)

// EditIssueComment edits the body of an existing issue/PR comment by its ID.
func (c *Client) EditIssueComment(ctx context.Context, owner, repo string, commentID int64, body string) error {
	_, _, err := c.gh.Issues.EditComment(ctx, owner, repo, commentID,
		&github.IssueComment{Body: github.Ptr(body)})
	return err
}

// DeleteIssueComment deletes an issue/PR comment by its ID.
func (c *Client) DeleteIssueComment(ctx context.Context, owner, repo string, commentID int64) error {
	_, err := c.gh.Issues.DeleteComment(ctx, owner, repo, commentID)
	return err
}

// LockConversation locks an issue or PR conversation. reason may be "off-topic",
// "too heated", "resolved", "spam", or "" for no reason.
func (c *Client) LockConversation(ctx context.Context, owner, repo string, number int, reason string) error {
	var opts *github.LockIssueOptions
	if reason != "" {
		opts = &github.LockIssueOptions{LockReason: reason}
	}
	_, err := c.gh.Issues.Lock(ctx, owner, repo, number, opts)
	return err
}

// UnlockConversation unlocks a previously locked issue or PR conversation.
func (c *Client) UnlockConversation(ctx context.Context, owner, repo string, number int) error {
	_, err := c.gh.Issues.Unlock(ctx, owner, repo, number)
	return err
}
