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

func (g *GitService) DiscardPath(path string) error {
	return gitops.Discard(g.repoDir, path)
}

func (g *GitService) Switch(branch string) error {
	return gitops.Switch(g.repoDir, branch)
}

func (g *GitService) SwitchNew(branch string) error {
	return gitops.SwitchNew(g.repoDir, branch)
}

func (g *GitService) DeleteBranch(name string, force bool) error {
	return gitops.DeleteBranch(g.repoDir, name, force)
}

func (g *GitService) RenameBranch(oldName, newName string) error {
	return gitops.RenameBranch(g.repoDir, oldName, newName)
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

func (g *GitService) Fetch() error {
	return gitops.Fetch(g.repoDir, "origin")
}

func (g *GitService) Pull(rebase bool) error {
	return gitops.Pull(g.repoDir, rebase)
}

func (g *GitService) Merge(branch string) error          { return gitops.Merge(g.repoDir, branch) }
func (g *GitService) MergeAbort() error                  { return gitops.MergeAbort(g.repoDir) }
func (g *GitService) ConflictedFiles() ([]string, error) { return gitops.ConflictedFiles(g.repoDir) }
func (g *GitService) MergeInProgress() bool              { return gitops.MergeInProgress(g.repoDir) }

func (g *GitService) Rebase(base string) error { return gitops.Rebase(g.repoDir, base) }
func (g *GitService) RebaseContinue() error    { return gitops.RebaseContinue(g.repoDir) }
func (g *GitService) RebaseAbort() error       { return gitops.RebaseAbort(g.repoDir) }
func (g *GitService) RebaseInProgress() bool   { return gitops.RebaseInProgress(g.repoDir) }

func (g *GitService) ReadConflict(path string) (*gitops.ConflictFile, error) {
	return gitops.ReadConflict(g.repoDir, path)
}

func (g *GitService) ResolveOurs(path string) error {
	return gitops.ResolveOurs(g.repoDir, path)
}

func (g *GitService) ResolveTheirs(path string) error {
	return gitops.ResolveTheirs(g.repoDir, path)
}

func (g *GitService) MarkResolved(path string) error {
	return gitops.MarkResolved(g.repoDir, path)
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

func (g *GitService) StashSave(message string, includeUntracked bool) error {
	return gitops.StashSave(g.repoDir, message, includeUntracked)
}

func (g *GitService) StashList() ([]gitops.Stash, error) {
	return gitops.StashList(g.repoDir)
}

func (g *GitService) StashPop(ref string) error {
	return gitops.StashPop(g.repoDir, ref)
}

func (g *GitService) StashDrop(ref string) error {
	return gitops.StashDrop(g.repoDir, ref)
}

func (g *GitService) Show(rev string) (*gitops.CommitDetail, error) {
	return gitops.Show(g.repoDir, rev)
}

func (g *GitService) CurrentBranch() (string, error) {
	return gitops.CurrentBranch(g.repoDir)
}

func (g *GitService) PRTemplate() string {
	return gitops.ReadPRTemplate(g.repoDir)
}

func (g *GitService) DefaultPRTitle(branch string) (string, error) {
	return gitops.LastCommitSubject(g.repoDir, branch)
}
