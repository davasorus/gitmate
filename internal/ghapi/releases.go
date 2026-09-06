package ghapi

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v88/github"
)

// Release is a trimmed GitHub release view.
type Release struct {
	ID         int64
	TagName    string
	Name       string
	Body       string
	Draft      bool
	Prerelease bool
	Immutable  bool
	URL        string
}

// Asset is a trimmed release-asset view.
type Asset struct {
	ID   int64
	Name string
	Size int
	URL  string
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

// CreateRelease creates a release on the given tag.
func (c *Client) CreateRelease(ctx context.Context, owner, repo, tag, name, body string, draft, prerelease bool) (Release, error) {
	r, _, err := c.gh.Repositories.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		TagName:    github.Ptr(tag),
		Name:       github.Ptr(name),
		Body:       github.Ptr(body),
		Draft:      github.Ptr(draft),
		Prerelease: github.Ptr(prerelease),
	})
	if err != nil {
		return Release{}, err
	}
	return toRelease(r), nil
}

// EditRelease updates an existing release by ID.
func (c *Client) EditRelease(ctx context.Context, owner, repo string, id int64, name, body string, draft, prerelease bool) (Release, error) {
	r, _, err := c.gh.Repositories.EditRelease(ctx, owner, repo, id, &github.RepositoryRelease{
		Name:       github.Ptr(name),
		Body:       github.Ptr(body),
		Draft:      github.Ptr(draft),
		Prerelease: github.Ptr(prerelease),
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

// GenerateReleaseNotes asks GitHub to auto-generate notes for a tag.
func (c *Client) GenerateReleaseNotes(ctx context.Context, owner, repo, tag string) (name, body string, err error) {
	notes, _, err := c.gh.Repositories.GenerateReleaseNotes(ctx, owner, repo, &github.GenerateNotesOptions{
		TagName: tag,
	})
	if err != nil {
		return "", "", err
	}
	return notes.Name, notes.Body, nil
}

// ListAssets returns the assets attached to a release.
func (c *Client) ListAssets(ctx context.Context, owner, repo string, releaseID int64) ([]Asset, error) {
	var out []Asset
	opts := &github.ListOptions{PerPage: 50}
	for {
		raw, resp, err := c.gh.Repositories.ListReleaseAssets(ctx, owner, repo, releaseID, opts)
		if err != nil {
			return nil, err
		}
		for _, a := range raw {
			out = append(out, Asset{ID: a.GetID(), Name: a.GetName(), Size: a.GetSize(), URL: a.GetBrowserDownloadURL()})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// UploadAsset attaches a file (name + raw bytes) to a release.
func (c *Client) UploadAsset(ctx context.Context, owner, repo string, releaseID int64, name string, data []byte) (Asset, error) {
	// go-github's UploadReleaseAsset handles the upload URL, content-length, and
	// content-type correctly (hand-rolling the raw request caused HTTP/2
	// PROTOCOL_ERRORs). It wants an *os.File, so stage the bytes in a temp file.
	tmp, err := os.CreateTemp("", "gitmate-asset-*")
	if err != nil {
		return Asset{}, err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return Asset{}, err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		_ = tmp.Close()
		return Asset{}, err
	}
	defer func() { _ = tmp.Close() }()

	asset, _, err := c.gh.Repositories.UploadReleaseAsset(ctx, owner, repo, releaseID,
		&github.UploadOptions{Name: name}, tmp)
	if err != nil {
		return Asset{}, err
	}
	return Asset{ID: asset.GetID(), Name: asset.GetName(), Size: asset.GetSize(), URL: asset.GetBrowserDownloadURL()}, nil
}

// DownloadAsset returns the raw bytes of a release asset.
func (c *Client) DownloadAsset(ctx context.Context, owner, repo string, assetID int64) ([]byte, error) {
	rc, _, err := c.gh.Repositories.DownloadReleaseAsset(ctx, owner, repo, assetID, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(rc)
}

// DeleteAsset removes an asset from a release.
func (c *Client) DeleteAsset(ctx context.Context, owner, repo string, assetID int64) error {
	_, err := c.gh.Repositories.DeleteReleaseAsset(ctx, owner, repo, assetID)
	return err
}

func toRelease(r *github.RepositoryRelease) Release {
	return Release{
		ID:         r.GetID(),
		TagName:    r.GetTagName(),
		Name:       r.GetName(),
		Body:       r.GetBody(),
		Draft:      r.GetDraft(),
		Prerelease: r.GetPrerelease(),
		Immutable:  r.GetImmutable(),
		URL:        r.GetHTMLURL(),
	}
}
