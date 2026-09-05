import { useEffect, useState, useCallback } from "react";
import { GitService } from "../bindings/github.com/davasorus/gitmate/gui";
import type { Status, Commit, Branch } from "../bindings/github.com/davasorus/gitmate/internal/gitops";
import type { PR } from "../bindings/github.com/davasorus/gitmate/internal/ghapi";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function App() {
  const [dir, setDir] = useState("");
  const [status, setStatus] = useState<Status | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [prs, setPRs] = useState<PR[]>([]);
  const [err, setErr] = useState("");

  const reload = useCallback(async () => {
    setErr("");
    try {
      await GitService.SetRepoDir(dir.trim());
      const [s, b, c] = await Promise.all([
        GitService.Status(),
        GitService.Branches(),
        GitService.Log(10),
      ]);
      setStatus(s);
      setBranches(b ?? []);
      setCommits(c ?? []);
      try {
        const p = await GitService.PRs("open");
        setPRs(p ?? []);
      } catch {
        setPRs([]); // no remote / no token — leave PRs empty, don't fail the page
      }
    } catch (e) {
      setErr(String(e));
    }
  }, [dir]);

  useEffect(() => { reload(); }, []); // initial load

  return (
    <div className="mx-auto max-w-4xl p-6 font-mono">
      <header className="mb-4 flex items-center justify-between border-b border-border pb-3">
        <h1 className="text-2xl font-bold">gitmate</h1>
        <div className="flex gap-2">
          <input
            value={dir}
            onChange={(e) => setDir(e.target.value)}
            placeholder="repo path (default .)"
            className="rounded-md border border-border bg-muted px-3 py-1.5 text-sm outline-none focus:border-primary"
          />
          <button onClick={reload} className="rounded-md bg-primary px-4 py-1.5 text-sm font-semibold text-background">
            Reload
          </button>
        </div>
      </header>

      {err && <div className="mb-4 text-sm text-red-500">{err}</div>}

      <Card className="mb-4">
        <CardHeader><CardTitle>Status</CardTitle></CardHeader>
        <CardContent>
          {status ? (
            <>
              <div className="mb-2">
                {status.Detached ? "HEAD detached" : <>On branch <b>{status.Branch}</b></>}
                {status.Upstream && <span className="text-muted-foreground"> — tracking {status.Upstream} (ahead {status.Ahead}, behind {status.Behind})</span>}
              </div>
              {((status.Changes?.length ?? 0) || (status.Untracked?.length ?? 0)) ? (
                <div className="space-y-1">
                  {(status.Changes ?? []).map((c, i) => (
                    <div key={i} className="flex gap-2 text-sm">
                      <Badge>{c.Staged}{c.Unstaged ? "/" + c.Unstaged : ""}</Badge>{c.Path}
                    </div>
                  ))}
                  {(status.Untracked ?? []).map((u, i) => (
                    <div key={i} className="flex gap-2 text-sm"><Badge>untracked</Badge>{u}</div>
                  ))}
                </div>
              ) : <span className="italic text-muted-foreground">working tree clean</span>}
            </>
          ) : "…"}
        </CardContent>
      </Card>

      <Card className="mb-4">
        <CardHeader><CardTitle>Branches</CardTitle></CardHeader>
        <CardContent className="space-y-1">
          {(branches ?? []).map((b) => (
            <div key={b.Name} className="flex items-baseline gap-2 text-sm">
              {b.IsCurrent && <span className="text-green-500">●</span>}
              <b>{b.Name}</b>
              <span className="text-amber-500">{b.LastHash}</span>
              {!b.Upstream && <span className="text-muted-foreground">(no upstream)</span>}
              {b.Upstream && (b.Ahead || b.Behind) ? <span className="text-muted-foreground">[ahead {b.Ahead}, behind {b.Behind}]</span> : null}
              <span className="truncate">{b.LastSubject}</span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card className="mb-4">
        <CardHeader><CardTitle>Recent commits</CardTitle></CardHeader>
        <CardContent className="space-y-1">
          {(commits ?? []).map((c) => (
            <div key={c.Hash} className="flex items-baseline gap-2 text-sm">
              <span className="text-amber-500">{c.Short}</span>
              <span className="truncate">{c.Subject}</span>
              <span className="text-muted-foreground">— {c.Author}</span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Open PRs</CardTitle></CardHeader>
        <CardContent className="space-y-1">
          {(prs ?? []).length ? (prs ?? []).map((p) => (
            <div key={p.Number} className="flex items-baseline gap-2 text-sm">
              <span className="font-semibold text-blue-500">#{p.Number}</span>
              <span className="truncate">{p.Title}</span>
              {p.Draft && <Badge variant="outline">draft</Badge>}
              <span className="text-muted-foreground">@{p.Author}</span>
            </div>
          )) : <span className="italic text-muted-foreground">no open PRs</span>}
        </CardContent>
      </Card>
    </div>
  );
}