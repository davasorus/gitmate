package ghapi

import (
	"context"

	"github.com/google/go-github/v75/github"
)

// Label is a repo label definition: name, hex color (no leading #), description.
type Label struct {
	Name        string
	Color       string
	Description string
}

// ListLabels returns all label definitions in the repo.
func (c *Client) ListLabels(ctx context.Context, owner, repo string) ([]Label, error) {
	var out []Label
	opts := &github.ListOptions{PerPage: 100}
	for {
		raw, resp, err := c.gh.Issues.ListLabels(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		for _, l := range raw {
			out = append(out, Label{Name: l.GetName(), Color: l.GetColor(), Description: l.GetDescription()})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// CreateLabel creates a new label definition (color = 6-hex, no #).
func (c *Client) CreateLabel(ctx context.Context, owner, repo, name, color, description string) error {
	_, _, err := c.gh.Issues.CreateLabel(ctx, owner, repo, &github.Label{
		Name:        github.String(name),
		Color:       github.String(color),
		Description: github.String(description),
	})
	return err
}

// EditLabel updates an existing label. newName can equal name to keep it.
func (c *Client) EditLabel(ctx context.Context, owner, repo, name, newName, color, description string) error {
	_, _, err := c.gh.Issues.EditLabel(ctx, owner, repo, name, &github.Label{
		Name:        github.String(newName),
		Color:       github.String(color),
		Description: github.String(description),
	})
	return err
}

// DeleteLabel removes a label definition from the repo.
func (c *Client) DeleteLabel(ctx context.Context, owner, repo, name string) error {
	_, err := c.gh.Issues.DeleteLabel(ctx, owner, repo, name)
	return err
}

// AddLabels adds labels to an issue or PR (PRs are issues here).
func (c *Client) AddLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	_, _, err := c.gh.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	return err
}

// RemoveLabel removes a single label from an issue or PR.
func (c *Client) RemoveLabel(ctx context.Context, owner, repo string, number int, label string) error {
	_, err := c.gh.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)
	return err
}

// SetLabels replaces all labels on an issue or PR with the given set.
func (c *Client) SetLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	_, _, err := c.gh.Issues.ReplaceLabelsForIssue(ctx, owner, repo, number, labels)
	return err
}
