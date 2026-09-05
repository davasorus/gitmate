package main

import (
	"context"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/davasorus/gitmate/internal/gitops"
)

// GitService is the bound service the frontend calls. Every method here
// delegates to the same internal/ engine the CLI uses.
type GitService struct {
	repoDir string // working directory for local git operations
}

// NewGitService starts pointed at the given directory ("." by default).
func NewGitService(dir string) *GitService {
	if dir == "" {
		dir = "."
	}
	return &GitService{repoDir: dir}
}

// SetRepoDir lets the frontend re-point the service at another repo.
func (g *GitService) SetRepoDir(dir string) {
	if dir != "" {
		g.repoDir = dir
	}
}

// GetRepoDir returns the current working directory.
func (g *GitService) GetRepoDir() string {
	return g.repoDir
}

// --- local git (reads .git, no network) ---

func (g *GitService) Status() (*gitops.Status, error) {
	return gitops.GetStatus(g.repoDir)
}

func (g *GitService) Log(limit int) ([]gitops.Commit, error) {
	return gitops.GetLog(g.repoDir, limit)
}

func (g *GitService) Branches() ([]gitops.Branch, error) {
	return gitops.GetBranches(g.repoDir)
}

func (g *GitService) Diff(path string, staged bool) ([]gitops.FileDiff, error) {
	return gitops.Diff(g.repoDir, gitops.DiffOptions{
		Path:   path,
		Staged: staged,
	})
}

// --- GitHub (needs GITHUB_TOKEN in the process env) ---

// PRs resolves owner/repo from the origin remote of the current dir.
func (g *GitService) PRs(state string) ([]ghapi.PR, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListPRs(ctx, owner, repo, state)
}

func (g *GitService) Issues(state string) ([]ghapi.Issue, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListIssues(ctx, owner, repo, state)
}

// resolve reads origin and parses owner/repo.
func (g *GitService) resolve(ctx context.Context) (owner, repo string, err error) {
	url, err := gitops.GetRemoteURL(g.repoDir, "origin")
	if err != nil {
		return "", "", err
	}
	return ghapi.ParseRepo(url)
}

// --- git write ---

func (g *GitService) Stage() error {
	return gitops.Stage(g.repoDir)
}

func (g *GitService) StagePath(path string) error {
	return gitops.Stage(g.repoDir, path)
}

func (g *GitService) UnstagePath(path string) error {
	return gitops.Unstage(g.repoDir, path)
}

func (g *GitService) Commit(message string) (string, error) {
	return gitops.CreateCommit(g.repoDir, message)
}

func (g *GitService) Push(setUpstream bool) error {
	branch, err := gitops.CurrentBranch(g.repoDir)
	if err != nil {
		return err
	}
	return gitops.Push(g.repoDir, "origin", branch, setUpstream)
}

// --- GitHub write ---

func (g *GitService) CreatePR(title, body, head, base string) (string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return "", err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	_, url, err := client.CreatePR(ctx, owner, repo, title, body, head, base)
	return url, err
}

func (g *GitService) CreateIssue(title, body string) (string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return "", err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	_, url, err := client.CreateIssue(ctx, owner, repo, title, body)
	return url, err
}

func (g *GitService) MergePR(number int, method string) (string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return "", err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return client.MergePR(ctx, owner, repo, number, method)
}

func (g *GitService) PRChecks(number int) ([]ghapi.CheckRun, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.PRChecks(ctx, owner, repo, number)
}
