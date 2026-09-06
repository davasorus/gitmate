package gitops

import "testing"

func TestReflog(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	entries, err := Reflog(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected reflog entries after a commit")
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 42: "42", -5: "-5"}
	for n, want := range cases {
		if got := itoa(n); got != want {
			t.Errorf("itoa(%d) = %q, want %q", n, got, want)
		}
	}
}
