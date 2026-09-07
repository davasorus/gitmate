package ghapi

import "context"

// AddAssignees adds users as assignees on an issue or pull request.
func (c *Client) AddAssignees(ctx context.Context, owner, repo string, number int, users []string) error {
	_, _, err := c.gh.Issues.AddAssignees(ctx, owner, repo, number, users)
	return err
}

// RemoveAssignees removes users from the assignees of an issue or pull request.
func (c *Client) RemoveAssignees(ctx context.Context, owner, repo string, number int, users []string) error {
	_, _, err := c.gh.Issues.RemoveAssignees(ctx, owner, repo, number, users)
	return err
}
