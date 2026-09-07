package main

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/davasorus/gitmate/internal/gitops"
)

// GitService is the bound service the frontend calls. Every method here
// delegates to the same internal/ engine the CLI uses.
type GitService struct {
	repoDir string // working directory for local git operations

	// lastUndo captures the state before the most recent history-moving op
	// (merge/rebase/reset/cherry-pick/revert), backing the "undo last" feature.
	lastUndo *undoPoint
}

// undoPoint is a restore point: the SHA HEAD pointed at before an operation,
// plus a human label for the affordance ("Merge feature-x").
type undoPoint struct {
	Label string // e.g. "Merge feature-x"
	SHA   string // HEAD before the op
}

// captureUndo records the current HEAD as the restore point for a labeled op.
// Best-effort: if HEAD can't be read (e.g. empty repo) it records nothing.
func (g *GitService) captureUndo(label string) {
	if sha, err := gitops.HeadSHA(g.repoDir); err == nil {
		g.lastUndo = &undoPoint{Label: label, SHA: sha}
	}
}

// UndoInfo describes the pending undoable operation for the UI (empty Label
// means nothing to undo).
type UndoInfo struct {
	Label string
	SHA   string
}

// LastUndoable returns the pending undo point (label + short sha), or an empty
// UndoInfo if there is nothing to undo.
func (g *GitService) LastUndoable() UndoInfo {
	if g.lastUndo == nil {
		return UndoInfo{}
	}
	short := g.lastUndo.SHA
	if len(short) > 7 {
		short = short[:7]
	}
	return UndoInfo{Label: g.lastUndo.Label, SHA: short}
}

// Undo hard-resets HEAD back to the captured restore point and clears it. The
// pre-undo state stays recoverable via the reflog.
func (g *GitService) Undo() error {
	if g.lastUndo == nil {
		return errors.New("nothing to undo")
	}
	if err := gitops.UndoTo(g.repoDir, g.lastUndo.SHA); err != nil {
		return err
	}
	g.lastUndo = nil
	return nil
}

// NewGitService starts pointed at the given directory ("." by default).
func NewGitService(dir string) *GitService {
	if dir == "" {
		dir = "."
	}
	// pin to the repo ROOT so path-relative git commands work regardless of the
	// process working dir (Wails runs from gui/).
	return &GitService{repoDir: gitops.RepoRoot(dir)}
}

// SetRepoDir lets the frontend re-point the service at another repo.
func (g *GitService) SetRepoDir(dir string) {
	if dir != "" {
		g.repoDir = gitops.RepoRoot(dir)
	}
}

// IsRepo reports whether the current repoDir is a git working tree.
func (g *GitService) IsRepo() bool {
	return gitops.IsRepo(g.repoDir)
}

// GetRepoDir returns the current working directory.
func (g *GitService) GetRepoDir() string {
	return g.repoDir
}

// --- local git (reads .git, no network) ---

// Status returns the working-tree status.
func (g *GitService) Status() (*gitops.Status, error) {
	return gitops.GetStatus(g.repoDir)
}

// Log returns recent commits.
func (g *GitService) Log(limit int) ([]gitops.Commit, error) {
	return gitops.GetLog(g.repoDir, limit)
}

// LogRef returns commits for a specific ref.
func (g *GitService) LogRef(ref string, limit int) ([]gitops.Commit, error) {
	return gitops.GetLogRef(g.repoDir, ref, limit)
}

// Branches returns local and remote-tracking branches.
func (g *GitService) Branches() ([]gitops.Branch, error) {
	return gitops.GetBranches(g.repoDir)
}

// Diff returns the diff for a path (staged or unstaged).
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

// PRsRich returns the PR list WITH review-decision + CI check rollup per PR,
// fetched in one GraphQL query (replaces the REST list + N per-PR check calls).
func (g *GitService) PRsRich(state string) ([]ghapi.PRListItem, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.PRListGraphQL(ctx, owner, repo, state)
}

// Issues returns issues by state (REST).
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

// IssuesRich returns issues WITH labels + assignees in one GraphQL query
// (and never mixes in PRs, unlike the REST issues endpoint).
func (g *GitService) IssuesRich(state string) ([]ghapi.IssueListItem, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.IssueListGraphQL(ctx, owner, repo, state)
}

// SetPRState opens or closes a pull request.
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

// SetIssueState opens or closes an issue.
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

// ListLabels returns the repository's labels.
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

// CreateLabel creates a label.
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

// EditLabel renames/recolors a label.
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

// DeleteLabel deletes a label.
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

// AddLabels adds labels to an issue or PR.
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

// RemoveLabel removes a label from an issue or PR.
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

// AddAssignees adds assignees to an issue or PR.
func (g *GitService) AddAssignees(number int, users []string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.AddAssignees(ctx, owner, repo, number, users)
}

// RemoveAssignees removes assignees from an issue or PR.
func (g *GitService) RemoveAssignees(number int, users []string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.RemoveAssignees(ctx, owner, repo, number, users)
}

// ListMilestones returns the repository's milestones by state.
func (g *GitService) ListMilestones(state string) ([]ghapi.Milestone, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListMilestones(ctx, owner, repo, state)
}

// SetMilestone assigns an issue/PR to a milestone (0 clears it).
func (g *GitService) SetMilestone(number, milestone int) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.SetMilestone(ctx, owner, repo, number, milestone)
}

// EditIssueComment edits an issue/PR comment by ID.
func (g *GitService) EditIssueComment(commentID int64, body string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.EditIssueComment(ctx, owner, repo, commentID, body)
}

// DeleteIssueComment deletes an issue/PR comment by ID.
func (g *GitService) DeleteIssueComment(commentID int64) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.DeleteIssueComment(ctx, owner, repo, commentID)
}

// LockConversation locks an issue/PR conversation (reason optional).
func (g *GitService) LockConversation(number int, reason string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.LockConversation(ctx, owner, repo, number, reason)
}

// UnlockConversation unlocks an issue/PR conversation.
func (g *GitService) UnlockConversation(number int) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.UnlockConversation(ctx, owner, repo, number)
}

// ListReleases returns the repository's releases.
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

// CreateRelease creates a release.
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

// EditRelease edits a release.
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

// DeleteRelease deletes a release.
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

// ListAssets returns a release's assets.
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

// DeleteAsset deletes a release asset.
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

// Stage stages all changes.
func (g *GitService) Stage() error {
	return gitops.Stage(g.repoDir)
}

// StagePath stages a specific path.
func (g *GitService) StagePath(path string) error {
	return gitops.Stage(g.repoDir, path)
}

// UnstagePath unstages a specific path.
// StageHunk stages a single hunk of a file (partial staging) by applying just
// that hunk's patch to the index.
func (g *GitService) StageHunk(path string, hunk gitops.Hunk) error {
	return gitops.StageHunk(g.repoDir, path, hunk)
}

// UnstageHunk removes a single staged hunk from the index.
func (g *GitService) UnstageHunk(path string, hunk gitops.Hunk) error {
	return gitops.UnstageHunk(g.repoDir, path, hunk)
}

func (g *GitService) UnstagePath(path string) error {
	return gitops.Unstage(g.repoDir, path)
}

// DiscardPath discards changes to a path.
func (g *GitService) DiscardPath(path string) error {
	return gitops.Discard(g.repoDir, path)
}

// Switch switches to a branch.
func (g *GitService) Switch(branch string) error {
	return gitops.Switch(g.repoDir, branch)
}

// SwitchNew creates and switches to a new branch.
func (g *GitService) SwitchNew(branch string) error {
	return gitops.SwitchNew(g.repoDir, branch)
}

// DeleteBranch deletes a branch.
func (g *GitService) DeleteBranch(name string, force bool) error {
	return gitops.DeleteBranch(g.repoDir, name, force)
}

// RenameBranch renames a branch.
func (g *GitService) RenameBranch(oldName, newName string) error {
	return gitops.RenameBranch(g.repoDir, oldName, newName)
}

// Commit commits staged changes.
func (g *GitService) Commit(message string) (string, error) {
	return gitops.CreateCommit(g.repoDir, message)
}

// Push pushes a branch to a remote.
func (g *GitService) Push(setUpstream bool) error {
	branch, err := gitops.CurrentBranch(g.repoDir)
	if err != nil {
		return err
	}
	return gitops.Push(g.repoDir, "origin", branch, setUpstream)
}

// Fetch fetches from origin (with prune).
func (g *GitService) Fetch() error {
	return gitops.Fetch(g.repoDir, "origin")
}

// Pull pulls from origin (merge or rebase).
func (g *GitService) Pull(rebase bool) error {
	return gitops.Pull(g.repoDir, rebase)
}

// Merge merges a branch into the current one.
func (g *GitService) Merge(branch string) error {
	g.captureUndo("Merge " + branch)
	return gitops.Merge(g.repoDir, branch)
}

// MergeAbort aborts an in-progress merge.
func (g *GitService) MergeAbort() error { return gitops.MergeAbort(g.repoDir) }

// ConflictedFiles returns the paths with merge conflicts.
func (g *GitService) ConflictedFiles() ([]string, error) { return gitops.ConflictedFiles(g.repoDir) }

// MergeInProgress reports whether a merge is in progress.
func (g *GitService) MergeInProgress() bool { return gitops.MergeInProgress(g.repoDir) }

// CommitMerge finishes an in-progress merge with git's prepared message.
func (g *GitService) CommitMerge() (string, error) { return gitops.CommitMerge(g.repoDir) }

// Rebase rebases the current branch onto a base.
func (g *GitService) Rebase(base string) error {
	g.captureUndo("Rebase onto " + base)
	return gitops.Rebase(g.repoDir, base)
}

// RebaseContinue resumes an in-progress rebase.
func (g *GitService) RebaseContinue() error { return gitops.RebaseContinue(g.repoDir) }

// RebaseAbort aborts an in-progress rebase.
func (g *GitService) RebaseAbort() error { return gitops.RebaseAbort(g.repoDir) }

// RebaseInProgress reports whether a rebase is in progress.
func (g *GitService) RebaseInProgress() bool { return gitops.RebaseInProgress(g.repoDir) }

// InteractiveRebaseTodo returns the commits from base..HEAD as a default
// interactive-rebase plan (all "pick"), for the UI to reorder/edit.
func (g *GitService) InteractiveRebaseTodo(base string) ([]gitops.RebaseStep, error) {
	return gitops.InteractiveRebaseTodo(g.repoDir, base)
}

// RunInteractiveRebase executes an interactive rebase onto base applying the
// given plan (reorder/drop/squash/fixup/reword). Captures an undo point first.
func (g *GitService) RunInteractiveRebase(base string, steps []gitops.RebaseStep) error {
	g.captureUndo("Interactive rebase onto " + base)
	return gitops.RunInteractiveRebase(g.repoDir, base, steps)
}

// ListTags returns tags (local + remote).
func (g *GitService) ListTags() ([]gitops.Tag, error) { return gitops.ListTags(g.repoDir) }

// CreateTag creates a tag.
func (g *GitService) CreateTag(name, message string) error {
	return gitops.CreateTag(g.repoDir, name, message)
}

// DeleteTag deletes a local tag.
func (g *GitService) DeleteTag(name string) error { return gitops.DeleteTag(g.repoDir, name) }

// PushTag pushes a tag to origin.
func (g *GitService) PushTag(name string) error { return gitops.PushTag(g.repoDir, name) }

// DeleteRemoteTag deletes a tag on origin.
func (g *GitService) DeleteRemoteTag(name string) error {
	return gitops.DeleteRemoteTag(g.repoDir, name)
}

// FetchTags fetches tags from origin.
func (g *GitService) FetchTags() error { return gitops.FetchTags(g.repoDir) }

// SmartDeleteTag deletes a tag locally and on origin.
func (g *GitService) SmartDeleteTag(name string) (string, error) {
	return gitops.SmartDeleteTag(g.repoDir, name)
}

// ReadConflict returns the conflict hunks for a file.
func (g *GitService) ReadConflict(path string) (*gitops.ConflictFile, error) {
	return gitops.ReadConflict(g.repoDir, path)
}

// ResolveOurs resolves a conflict by taking our side.
func (g *GitService) ResolveOurs(path string) error {
	return gitops.ResolveOurs(g.repoDir, path)
}

// ResolveTheirs resolves a conflict by taking their side.
func (g *GitService) ResolveTheirs(path string) error {
	return gitops.ResolveTheirs(g.repoDir, path)
}

// MarkResolved marks a conflicted file resolved.
func (g *GitService) MarkResolved(path string) error {
	return gitops.MarkResolved(g.repoDir, path)
}

// --- GitHub write ---

// CreatePR opens a pull request.
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

// CreateIssue opens an issue.
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

// MergePR merges a pull request.
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

// PRChecks returns the CI check runs for a PR's head commit.
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

// ListReviews returns a PR's reviews.
func (g *GitService) ListReviews(number int) ([]ghapi.Review, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListReviews(ctx, owner, repo, number)
}

// PRDiff returns a PR's unified diff.
func (g *GitService) PRDiff(number int) ([]gitops.FileDiff, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.PRDiff(ctx, owner, repo, number)
}

// ListReviewComments returns a PR's review (line) comments.
func (g *GitService) ListReviewComments(number int) ([]ghapi.ExistingComment, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListReviewComments(ctx, owner, repo, number)
}

// ListIssueComments returns a PR's general issue-comment stream.
func (g *GitService) ListIssueComments(number int) ([]ghapi.IssueComment, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListIssueComments(ctx, owner, repo, number)
}

// ReplyToReviewComment replies to a review comment.
func (g *GitService) ReplyToReviewComment(number int, commentID int64, body string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.ReplyToReviewComment(ctx, owner, repo, number, commentID, body)
}

// PRDetail returns the aggregated PR detail (reviews, threads, checks, labels, assignees) via GraphQL.
func (g *GitService) PRDetail(number int) (*ghapi.PRDetail, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.PRDetailGraphQL(ctx, owner, repo, number)
}

// ListRuns returns recent workflow runs.
func (g *GitService) ListRuns(limit int) ([]ghapi.WorkflowRun, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListRuns(ctx, owner, repo, limit)
}

// RunJobs returns a run's jobs.
func (g *GitService) RunJobs(runID int64) ([]ghapi.Job, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListRunJobs(ctx, owner, repo, runID)
}

// GetRun returns a single run.
func (g *GitService) GetRun(runID int64) (ghapi.WorkflowRun, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return ghapi.WorkflowRun{}, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return ghapi.WorkflowRun{}, err
	}
	return client.GetRun(ctx, owner, repo, runID)
}

// CancelRun cancels a run (needs workflow scope).
func (g *GitService) CancelRun(runID int64) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.CancelRun(ctx, owner, repo, runID)
}

// RerunRun reruns a run (needs workflow scope).
func (g *GitService) RerunRun(runID int64) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.RerunRun(ctx, owner, repo, runID)
}

// RerunFailed reruns a run's failed jobs (needs workflow scope).
func (g *GitService) RerunFailed(runID int64) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.RerunFailed(ctx, owner, repo, runID)
}

// JobLogs returns a job's logs, split by step.
func (g *GitService) JobLogs(jobID int64) (ghapi.JobLog, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return ghapi.JobLog{}, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return ghapi.JobLog{}, err
	}
	return client.JobLogs(ctx, owner, repo, jobID)
}

// RunJobGraph returns the run's job dependency graph (matrix legs expanded).
func (g *GitService) RunJobGraph(runID int64) ([]ghapi.JobNode, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.RunJobGraph(ctx, owner, repo, runID)
}

// ListDispatchableWorkflows returns workflows that accept workflow_dispatch, with their inputs.
func (g *GitService) ListDispatchableWorkflows() ([]ghapi.DispatchableWorkflow, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListDispatchableWorkflows(ctx, owner, repo)
}

// TriggerDispatch fires a workflow_dispatch. inputs is a JSON-ish string map from
// the frontend; values pass through as strings (GitHub accepts string inputs).
func (g *GitService) TriggerDispatch(workflowFile, ref string, inputs map[string]string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	m := make(map[string]interface{}, len(inputs))
	for k, v := range inputs {
		m[k] = v
	}
	return client.TriggerDispatch(ctx, owner, repo, workflowFile, ref, m)
}

// ResolveThread marks a review thread resolved (GraphQL).
func (g *GitService) ResolveThread(threadID string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.ResolveThread(ctx, threadID)
}

// UnresolveThread reopens a resolved review thread (GraphQL).
func (g *GitService) UnresolveThread(threadID string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.UnresolveThread(ctx, threadID)
}

// CommentPR adds a general comment to a PR.
func (g *GitService) CommentPR(number int, body string) (string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return "", err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return client.CommentPR(ctx, owner, repo, number, body)
}

// SubmitReview submits a PR review (approve/request-changes/comment).
func (g *GitService) SubmitReview(number int, event, body string, comments []ghapi.ReviewComment) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.SubmitReview(ctx, owner, repo, number, event, body, comments)
}

// ListRequestedReviewers returns a PR's requested reviewers.
func (g *GitService) ListRequestedReviewers(number int) ([]ghapi.Reviewer, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return nil, err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return client.ListRequestedReviewers(ctx, owner, repo, number)
}

// RequestReviewers requests reviewers on a PR.
func (g *GitService) RequestReviewers(number int, logins []string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.RequestReviewers(ctx, owner, repo, number, logins)
}

// RemoveReviewer removes a requested reviewer from a PR.
func (g *GitService) RemoveReviewer(number int, login string) error {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.RemoveReviewer(ctx, owner, repo, number, login)
}

// StashSave saves a stash.
func (g *GitService) StashSave(message string, includeUntracked bool) error {
	return gitops.StashSave(g.repoDir, message, includeUntracked)
}

// StashList returns the stash list.
func (g *GitService) StashList() ([]gitops.Stash, error) {
	return gitops.StashList(g.repoDir)
}

// StashPop pops a stash.
func (g *GitService) StashPop(ref string) error {
	return gitops.StashPop(g.repoDir, ref)
}

// StashApply applies a stash without dropping it.
func (g *GitService) StashApply(ref string) error {
	return gitops.StashApply(g.repoDir, ref)
}

// StashDrop drops a stash.
func (g *GitService) StashDrop(ref string) error {
	return gitops.StashDrop(g.repoDir, ref)
}

// Show returns a commit's detail + diff.
func (g *GitService) Show(rev string) (*gitops.CommitDetail, error) {
	return gitops.Show(g.repoDir, rev)
}

// Reflog returns reflog entries.
func (g *GitService) Reflog(limit int) ([]gitops.ReflogEntry, error) {
	return gitops.Reflog(g.repoDir, limit)
}

// Blame returns line-by-line blame for a file.
func (g *GitService) Blame(path string) ([]gitops.BlameLine, error) {
	return gitops.Blame(g.repoDir, path)
}

// ListRemotes returns configured remotes.
func (g *GitService) ListRemotes() ([]gitops.Remote, error) { return gitops.ListRemotes(g.repoDir) }

// AddRemote adds a remote.
func (g *GitService) AddRemote(name, url string) error { return gitops.AddRemote(g.repoDir, name, url) }

// RemoveRemote removes a remote.
func (g *GitService) RemoveRemote(name string) error { return gitops.RemoveRemote(g.repoDir, name) }

// RenameRemote renames a remote.
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

// Reset resets HEAD to a ref (soft/mixed/hard).
func (g *GitService) Reset(rev, mode string) error {
	g.captureUndo("Reset (" + mode + ") to " + rev)
	return gitops.Reset(g.repoDir, rev, gitops.ResetMode(mode))
}

// CherryPick cherry-picks a commit.
func (g *GitService) CherryPick(rev string) error {
	g.captureUndo("Cherry-pick " + rev)
	return gitops.CherryPick(g.repoDir, rev)
}

// CherryPickContinue resumes an in-progress cherry-pick after conflicts are resolved.
func (g *GitService) CherryPickContinue() error { return gitops.CherryPickContinue(g.repoDir) }

// CherryPickAbort aborts an in-progress cherry-pick.
func (g *GitService) CherryPickAbort() error { return gitops.CherryPickAbort(g.repoDir) }

// Revert reverts a commit.
func (g *GitService) Revert(rev string) error {
	g.captureUndo("Revert " + rev)
	return gitops.Revert(g.repoDir, rev)
}

// RevertContinue resumes an in-progress revert after conflicts are resolved.
func (g *GitService) RevertContinue() error { return gitops.RevertContinue(g.repoDir) }

// RevertAbort aborts an in-progress revert.
func (g *GitService) RevertAbort() error { return gitops.RevertAbort(g.repoDir) }

// SequencerInProgress reports cherry-pick/revert in-progress state.
func (g *GitService) SequencerInProgress() (bool, bool) { return gitops.SequencerInProgress(g.repoDir) }

// CurrentBranch returns the current branch name.
func (g *GitService) CurrentBranch() (string, error) {
	return gitops.CurrentBranch(g.repoDir)
}

// PRTemplate returns the repo's PR template body, if any.
func (g *GitService) PRTemplate() string {
	return gitops.ReadPRTemplate(g.repoDir)
}

// DefaultPRTitle suggests a default PR title from the branch.
func (g *GitService) DefaultPRTitle(branch string) (string, error) {
	return gitops.LastCommitSubject(g.repoDir, branch)
}

// DefaultBranch returns the repo default branch (base for new PRs).
func (g *GitService) DefaultBranch() (string, error) {
	ctx := context.Background()
	owner, repo, err := g.resolve(ctx)
	if err != nil {
		return "", err
	}
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return client.DefaultBranch(ctx, owner, repo)
}
