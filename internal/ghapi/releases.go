package ghapi

import (
	"context"

	"github.com/google/go-github/v66/github"
)

// Release is a trimmed GitHub release view.
type Release struct {
	ID         int64
	TagName    string
	Name       string
	Body       string
	Draft      bool
	Prerelease bool
	URL        string // HTML URL
}

// ListReleases returns the repo's releases, newest first.
func (c *Client) ListReleases(ctx context.Context, owner, repo string) ([]Release, error) {
	var out []Release
	opts := &github.ListOptions{PerPage: 30}
	for {
		raw, resp, err := c.gh.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		for _, r := range raw {
			out = append(out, toRelease(r))
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// CreateRelease creates a release on the given tag. If the tag does not exist,
// GitHub creates it at the default branch's head (target can override — left as
// default here). draft/prerelease control visibility/stability flags.
func (c *Client) CreateRelease(ctx context.Context, owner, repo, tag, name, body string, draft, prerelease bool) (Release, error) {
	r, _, err := c.gh.Repositories.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		TagName:    github.String(tag),
		Name:       github.String(name),
		Body:       github.String(body),
		Draft:      github.Bool(draft),
		Prerelease: github.Bool(prerelease),
	})
	if err != nil {
		return Release{}, err
	}
	return toRelease(r), nil
}

// EditRelease updates an existing release by ID.
func (c *Client) EditRelease(ctx context.Context, owner, repo string, id int64, name, body string, draft, prerelease bool) (Release, error) {
	r, _, err := c.gh.Repositories.EditRelease(ctx, owner, repo, id, &github.RepositoryRelease{
		Name:       github.String(name),
		Body:       github.String(body),
		Draft:      github.Bool(draft),
		Prerelease: github.Bool(prerelease),
	})
	if err != nil {
		return Release{}, err
	}
	return toRelease(r), nil
}

// DeleteRelease removes a release by ID (does NOT delete the underlying tag).
func (c *Client) DeleteRelease(ctx context.Context, owner, repo string, id int64) error {
	_, err := c.gh.Repositories.DeleteRelease(ctx, owner, repo, id)
	return err
}

// GenerateReleaseNotes asks GitHub to auto-generate notes for a tag (from merged
// PRs / commits since the previous tag). Returns the generated name + body.
func (c *Client) GenerateReleaseNotes(ctx context.Context, owner, repo, tag string) (name, body string, err error) {
	notes, _, err := c.gh.Repositories.GenerateReleaseNotes(ctx, owner, repo, &github.GenerateNotesOptions{
		TagName: tag,
	})
	if err != nil {
		return "", "", err
	}
	return notes.Name, notes.Body, nil
}

func toRelease(r *github.RepositoryRelease) Release {
	return Release{
		ID:         r.GetID(),
		TagName:    r.GetTagName(),
		Name:       r.GetName(),
		Body:       r.GetBody(),
		Draft:      r.GetDraft(),
		Prerelease: r.GetPrerelease(),
		URL:        r.GetHTMLURL(),
	}
}
