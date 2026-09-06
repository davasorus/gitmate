package main

import (
	"context"
	"encoding/base64"

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

func (g *GitService) LogRef(ref string, limit int) ([]gitops.Commit, error) {
	return gitops.GetLogRef(g.repoDir, ref, limit)
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

func (g *GitService) SetPRState(number int, state string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.SetPRState(ctx, owner, repo, number, state)
}

func (g *GitService) SetIssueState(number int, state string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.SetIssueState(ctx, owner, repo, number, state)
}

func (g *GitService) ListLabels() ([]ghapi.Label, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListLabels(ctx, owner, repo)
}

func (g *GitService) CreateLabel(name, color, description string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.CreateLabel(ctx, owner, repo, name, color, description)
}

func (g *GitService) EditLabel(name, newName, color, description string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.EditLabel(ctx, owner, repo, name, newName, color, description)
}

func (g *GitService) DeleteLabel(name string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.DeleteLabel(ctx, owner, repo, name)
}

func (g *GitService) AddLabels(number int, labels []string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.AddLabels(ctx, owner, repo, number, labels)
}

func (g *GitService) RemoveLabel(number int, label string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.RemoveLabel(ctx, owner, repo, number, label)
}

func (g *GitService) ListReleases() ([]ghapi.Release, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListReleases(ctx, owner, repo)
}

func (g *GitService) CreateRelease(tag, name, body string, draft, prerelease bool) (ghapi.Release, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return ghapi.Release{}, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return ghapi.Release{}, err
	}
	return client.CreateRelease(ctx, owner, repo, tag, name, body, draft, prerelease)
}

func (g *GitService) EditRelease(id int64, name, body string, draft, prerelease bool) (ghapi.Release, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return ghapi.Release{}, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return ghapi.Release{}, err
	}
	return client.EditRelease(ctx, owner, repo, id, name, body, draft, prerelease)
}

func (g *GitService) DeleteRelease(id int64) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.DeleteRelease(ctx, owner, repo, id)
}

func (g *GitService) ListAssets(releaseID int64) ([]ghapi.Asset, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListAssets(ctx, owner, repo, releaseID)
}

// UploadAsset takes the file name and its base64-encoded contents (from the
// frontend file input) and attaches it to the release.
func (g *GitService) UploadAsset(releaseID int64, name, dataB64 string) (ghapi.Asset, error) {
	data, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return ghapi.Asset{}, err
	}
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return ghapi.Asset{}, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return ghapi.Asset{}, err
	}
	return client.UploadAsset(ctx, owner, repo, releaseID, name, data)
}

// DownloadAsset returns the asset's contents base64-encoded so the frontend can
// trigger a browser download.
func (g *GitService) DownloadAsset(assetID int64) (string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return "", err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	data, err := client.DownloadAsset(ctx, owner, repo, assetID)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (g *GitService) DeleteAsset(assetID int64) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.DeleteAsset(ctx, owner, repo, assetID)
}

// GenerateReleaseNotes returns [name, body] so it binds cleanly to TS.
func (g *GitService) GenerateReleaseNotes(tag string) ([]string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	name, body, err := client.GenerateReleaseNotes(ctx, owner, repo, tag)
	if err != nil {
		return nil, err
	}
	return []string{name, body}, nil
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

func (g *GitService) ListTags() ([]gitops.Tag, error) { return gitops.ListTags(g.repoDir) }
func (g *GitService) CreateTag(name, message string) error {
	return gitops.CreateTag(g.repoDir, name, message)
}
func (g *GitService) DeleteTag(name string) error { return gitops.DeleteTag(g.repoDir, name) }
func (g *GitService) PushTag(name string) error   { return gitops.PushTag(g.repoDir, name) }

func (g *GitService) DeleteRemoteTag(name string) error {
	return gitops.DeleteRemoteTag(g.repoDir, name)
}
func (g *GitService) FetchTags() error { return gitops.FetchTags(g.repoDir) }
func (g *GitService) SmartDeleteTag(name string) (string, error) {
	return gitops.SmartDeleteTag(g.repoDir, name)
}

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

func (g *GitService) StashApply(ref string) error {
	return gitops.StashApply(g.repoDir, ref)
}

func (g *GitService) StashDrop(ref string) error {
	return gitops.StashDrop(g.repoDir, ref)
}

func (g *GitService) Show(rev string) (*gitops.CommitDetail, error) {
	return gitops.Show(g.repoDir, rev)
}

func (g *GitService) Reflog(limit int) ([]gitops.ReflogEntry, error) {
	return gitops.Reflog(g.repoDir, limit)
}

func (g *GitService) Blame(path string) ([]gitops.BlameLine, error) {
	return gitops.Blame(g.repoDir, path)
}

func (g *GitService) ListRemotes() ([]gitops.Remote, error) { return gitops.ListRemotes(g.repoDir) }
func (g *GitService) AddRemote(name, url string) error      { return gitops.AddRemote(g.repoDir, name, url) }
func (g *GitService) RemoveRemote(name string) error        { return gitops.RemoveRemote(g.repoDir, name) }
func (g *GitService) RenameRemote(oldName, newName string) error {
	return gitops.RenameRemote(g.repoDir, oldName, newName)
}

// Clone clones url into dest and points the service at the new repo so the app
// switches to it. Returns the cloned repo's path.
func (g *GitService) Clone(url, dest string) (string, error) {
	path, err := gitops.Clone(url, dest)
	if err != nil {
		return "", err
	}
	g.repoDir = path
	return path, nil
}

func (g *GitService) Reset(rev, mode string) error {
	return gitops.Reset(g.repoDir, rev, gitops.ResetMode(mode))
}

func (g *GitService) CherryPick(rev string) error       { return gitops.CherryPick(g.repoDir, rev) }
func (g *GitService) CherryPickContinue() error         { return gitops.CherryPickContinue(g.repoDir) }
func (g *GitService) CherryPickAbort() error            { return gitops.CherryPickAbort(g.repoDir) }
func (g *GitService) Revert(rev string) error           { return gitops.Revert(g.repoDir, rev) }
func (g *GitService) RevertContinue() error             { return gitops.RevertContinue(g.repoDir) }
func (g *GitService) RevertAbort() error                { return gitops.RevertAbort(g.repoDir) }
func (g *GitService) SequencerInProgress() (bool, bool) { return gitops.SequencerInProgress(g.repoDir) }

func (g *GitService) CurrentBranch() (string, error) {
	return gitops.CurrentBranch(g.repoDir)
}

func (g *GitService) PRTemplate() string {
	return gitops.ReadPRTemplate(g.repoDir)
}

func (g *GitService) DefaultPRTitle(branch string) (string, error) {
	return gitops.LastCommitSubject(g.repoDir, branch)
}
