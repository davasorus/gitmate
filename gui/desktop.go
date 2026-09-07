package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// SelectDirectory opens the native OS folder picker and returns the chosen path
// (empty string if the user cancels). Used by the "Open repository" flow.
func (g *GitService) SelectDirectory() string {
	path, err := application.Get().Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("Open repository").
		PromptForSingleSelection()
	if err != nil {
		return ""
	}
	return path
}

// --- recent repositories -------------------------------------------------
// Persisted as a small JSON list in the user's config dir, so recently opened
// repos are one click away on relaunch.

const maxRecent = 8

func recentPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "gitmate", "recent.json")
}

// RecentRepos returns the recently opened repository paths, most recent first.
func (g *GitService) RecentRepos() []string {
	p := recentPath()
	if p == "" {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var out []string
	if json.Unmarshal(data, &out) != nil {
		return nil
	}
	return out
}

// AddRecentRepo records a repo path as most-recently-opened (deduped, capped).
func (g *GitService) AddRecentRepo(path string) {
	if path == "" {
		return
	}
	p := recentPath()
	if p == "" {
		return
	}
	list := g.RecentRepos()
	// move-to-front, dedupe
	next := []string{path}
	for _, x := range list {
		if x != path && len(next) < maxRecent {
			next = append(next, x)
		}
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if data, err := json.Marshal(next); err == nil {
		_ = os.WriteFile(p, data, 0o644)
	}
}

// --- persisted settings --------------------------------------------------
// The repository directory is a first-class, persisted setting — the source of
// truth lives here, not in transient frontend state. Loaded once at startup,
// saved whenever the user opens a different repo.

type appSettings struct {
	RepoDir string `json:"repoDir"`
}

func settingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "gitmate", "settings.json")
}

func loadSettings() appSettings {
	var s appSettings
	p := settingsPath()
	if p == "" {
		return s
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	return s
}

func saveSettings(s appSettings) {
	p := settingsPath()
	if p == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if data, err := json.Marshal(s); err == nil {
		_ = os.WriteFile(p, data, 0o644)
	}
}

// RepoDir returns the current repository directory (the persisted setting).
func (g *GitService) RepoDir() string {
	return g.repoDir
}
