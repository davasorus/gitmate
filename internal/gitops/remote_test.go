package gitops

import "testing"

func TestListRenameRemoveRemote(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	if err := AddRemote(dir, "origin", "https://github.com/o/r.git"); err != nil {
		t.Fatal(err)
	}
	rs, err := ListRemotes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) == 0 || rs[0].Name != "origin" {
		t.Fatalf("remotes wrong: %+v", rs)
	}
	if err := RenameRemote(dir, "origin", "upstream"); err != nil {
		t.Fatal(err)
	}
	if rs, _ = ListRemotes(dir); rs[0].Name != "upstream" {
		t.Fatalf("rename failed: %+v", rs)
	}
	if err := RemoveRemote(dir, "upstream"); err != nil {
		t.Fatal(err)
	}
	if rs, _ = ListRemotes(dir); len(rs) != 0 {
		t.Fatalf("expected no remotes: %+v", rs)
	}
}
