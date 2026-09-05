package gitops

import "strings"

// BlameLine is one line of a file with the commit that last touched it.
type BlameLine struct {
	Line    int    // 1-based line number in the current file
	Short   string // short commit hash
	Author  string // author name
	Content string // the line's text
}

// Blame returns per-line authorship for a file using `git blame --porcelain`,
// which emits stable, machine-readable records (one header block per line group
// plus a tab-prefixed content line).
func Blame(dir, path string) ([]BlameLine, error) {
	out, err := run(dir, "blame", "--porcelain", "--", path)
	if err != nil {
		return nil, err
	}

	var lines []BlameLine
	// porcelain format: a header line "<sha> <origLine> <finalLine> [<group>]"
	// followed by key/value lines (author, ...), then the content line prefixed
	// with a TAB. Author names carry forward within a commit's group.
	authorBySha := map[string]string{}
	var curSha, curAuthor string
	var curFinal int

	for _, raw := range strings.Split(out, "\n") {
		if raw == "" {
			continue
		}
		if raw[0] == '\t' {
			// content line for the current header
			lines = append(lines, BlameLine{
				Line:    curFinal,
				Short:   shortSha(curSha),
				Author:  curAuthor,
				Content: raw[1:], // strip the leading tab
			})
			continue
		}
		// header or metadata
		if strings.HasPrefix(raw, "author ") {
			curAuthor = strings.TrimPrefix(raw, "author ")
			if curSha != "" {
				authorBySha[curSha] = curAuthor
			}
			continue
		}
		// a header line starts with 40-hex sha and space-separated numbers
		fields := strings.Fields(raw)
		if len(fields) >= 3 && len(fields[0]) >= 7 && isHex(fields[0]) {
			curSha = fields[0]
			// final line number is the 3rd field
			curFinal = atoiSafe(fields[2])
			// reuse remembered author if we've seen this sha before
			if a, ok := authorBySha[curSha]; ok {
				curAuthor = a
			}
		}
	}
	return lines, nil
}

func shortSha(s string) string {
	if len(s) >= 7 {
		return s[:7]
	}
	return s
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) > 0
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
