package gitops

import "testing"

func TestBlame(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	lines, err := Blame(dir, "a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Fatal("expected blame lines")
	}
}

func TestBlameHelpers(t *testing.T) {
	// shortSha: long → truncated, short → unchanged
	if shortSha("abcdef1234567") != "abcdef1" {
		t.Errorf("shortSha long wrong")
	}
	if shortSha("abc") != "abc" {
		t.Errorf("shortSha short wrong")
	}
	// isHex: valid, invalid, empty
	if !isHex("deadBEEF01") {
		t.Errorf("isHex valid failed")
	}
	if isHex("xyz") {
		t.Errorf("isHex should reject non-hex")
	}
	if isHex("") {
		t.Errorf("isHex should reject empty")
	}
	// atoiSafe: digits, leading digits then junk, non-digit
	if atoiSafe("123") != 123 {
		t.Errorf("atoiSafe digits wrong")
	}
	if atoiSafe("45abc") != 45 {
		t.Errorf("atoiSafe should stop at non-digit")
	}
	if atoiSafe("nope") != 0 {
		t.Errorf("atoiSafe non-digit should be 0")
	}
}
