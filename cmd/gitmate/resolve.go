package main

import (
	"errors"
	"strings"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/davasorus/gitmate/internal/gitops"
)

// resolveRepo figures out owner/repo: explicit --repo flag wins,
// otherwise parse the origin remote of the current directory.
func resolveRepo(flag string) (owner, repo string, err error) {
	if flag != "" {
		parts := strings.Split(flag, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", errors.New("--repo must be in the form owner/name")
		}
		return parts[0], parts[1], nil
	}
	url, err := gitops.GetRemoteURL(".", "origin")
	if err != nil {
		return "", "", errors.New("no origin remote found; pass --repo owner/name")
	}
	return ghapi.ParseRepo(url)
}
