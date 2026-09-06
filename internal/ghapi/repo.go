package ghapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v88/github"
)

// ParseRepo extracts owner/repo from a git remote URL in either form:
//
//	https://github.com/owner/repo.git
//	git@github.com:owner/repo.git
func ParseRepo(url string) (owner, repo string, err error) {
	u := strings.TrimSpace(url)
	u = strings.TrimSuffix(u, ".git")

	switch {
	case strings.HasPrefix(u, "https://github.com/"):
		u = strings.TrimPrefix(u, "https://github.com/")
	case strings.HasPrefix(u, "git@github.com:"):
		u = strings.TrimPrefix(u, "git@github.com:")
	default:
		return "", "", fmt.Errorf("unrecognized GitHub remote URL: %s", url)
	}

	parts := strings.Split(u, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("could not parse owner/repo from: %s", url)
	}
	return parts[0], parts[1], nil
}

// CreateRepo creates a new repository under the authenticated user's
// account. Passing "" as the org means "create it on my own account."
func (c *Client) CreateRepo(ctx context.Context, name, description string, private bool) (string, error) {
	repo := &github.Repository{
		Name:        github.Ptr(name),
		Description: github.Ptr(description),
		Private:     github.Ptr(private),
	}
	created, _, err := c.gh.Repositories.Create(ctx, "", repo)
	if err != nil {
		return "", err
	}
	return created.GetCloneURL(), nil // https://github.com/owner/name.git
}
