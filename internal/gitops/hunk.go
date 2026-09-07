package gitops

import (
	"strings"
)

// buildHunkPatch reconstructs a minimal, apply-able unified-diff patch containing
// a SINGLE hunk for one file. It reuses the hunk's original @@ header verbatim
// (so line counts are exact — never recomputed) and re-prefixes each line by its
// kind. The trailing newline matters: git apply is strict about patch format.
func buildHunkPatch(path string, h Hunk) string {
	var b strings.Builder
	b.WriteString("diff --git a/" + path + " b/" + path + "\n")
	b.WriteString("--- a/" + path + "\n")
	b.WriteString("+++ b/" + path + "\n")
	b.WriteString(h.Header + "\n")
	for _, ln := range h.Lines {
		switch ln.Kind {
		case LineAdd:
			b.WriteString("+" + ln.Content + "\n")
		case LineRemove:
			b.WriteString("-" + ln.Content + "\n")
		default: // context
			b.WriteString(" " + ln.Content + "\n")
		}
	}
	return b.String()
}

// StageHunk stages a single hunk of a file by building a patch for just that
// hunk and applying it to the index (git apply --cached). This is how you stage
// part of a file's changes without staging the whole file.
func StageHunk(dir, path string, h Hunk) error {
	patch := buildHunkPatch(path, h)
	return applyPatch(dir, patch, false)
}

// UnstageHunk removes a single staged hunk from the index by applying its patch
// in reverse against the index (git apply --cached --reverse).
func UnstageHunk(dir, path string, h Hunk) error {
	patch := buildHunkPatch(path, h)
	return applyPatch(dir, patch, true)
}

// applyPatch feeds a patch to `git apply --cached` (optionally --reverse) via
// stdin. --cached applies to the index only, leaving the working tree untouched.
func applyPatch(dir, patch string, reverse bool) error {
	args := []string{"apply", "--cached", "--unidiff-zero"}
	if reverse {
		args = append(args, "--reverse")
	}
	args = append(args, "-") // read patch from stdin
	return runStdin(dir, patch, args...)
}
