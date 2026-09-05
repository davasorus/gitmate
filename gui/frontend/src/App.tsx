import { useEffect, useState, useCallback } from "react";
import { GitService } from "../bindings/github.com/davasorus/gitmate/gui";
import type { Status, Commit, Branch } from "../bindings/github.com/davasorus/gitmate/internal/gitops";
import type { PR, CheckRun } from "../bindings/github.com/davasorus/gitmate/internal/ghapi";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

type Toast = { kind: "ok" | "err"; msg: string } | null;

export default function App() {
  const [dir, setDir] = useState("");
  const [status, setStatus] = useState<Status | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [prs, setPRs] = useState<PR[]>([]);
  const [toast, setToast] = useState<Toast>(null);
  const [busy, setBusy] = useState<string>("");

  // per-PR check results, keyed by PR number
  const [checks, setChecks] = useState<Record<number, CheckRun[]>>({});

  // form state
  const [commitMsg, setCommitMsg] = useState("");
  const [prTitle, setPrTitle] = useState("");
  const [prHead, setPrHead] = useState("");
  const [prBase, setPrBase] = useState("live");
  const [issueTitle, setIssueTitle] = useState("");

  const flash = (kind: "ok" | "err", msg: string) => {
    setToast({ kind, msg });
    setTimeout(() => setToast(null), 4000);
  };

  const reload = useCallback(async () => {
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
        setPRs([]);
      }
    } catch (e) {
      flash("err", String(e));
    }
  }, [dir]);

  useEffect(() => { reload(); }, []);

  const run = async (name: string, fn: () => Promise<string | void>, okMsg: string) => {
    setBusy(name);
    try {
      const res = await fn();
      flash("ok", typeof res === "string" && res ? res : okMsg);
      await reload();
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const doCommit = () =>
    run("commit", async () => {
      await GitService.Stage();
      const hash = await GitService.Commit(commitMsg);
      setCommitMsg("");
      return `committed ${hash}`;
    }, "committed");

  const doPush = () => run("push", () => GitService.Push(true), "pushed");
  const doPR = () => run("pr", () => GitService.CreatePR(prTitle, "", prHead, prBase), "PR opened");
  const doIssue = () => run("issue", () => GitService.CreateIssue(issueTitle, ""), "issue opened");

  const doMerge = (n: number) =>
    run(`merge-${n}`, async () => {
      const sha = await GitService.MergePR(n, "merge");
      return `merged #${n} (${sha.slice(0, 7)})`;
    }, `merged #${n}`);

  // Load checks for one PR on demand (doesn't trigger a full reload).
  const loadChecks = async (n: number) => {
    setBusy(`checks-${n}`);
    try {
      const runs = await GitService.PRChecks(n);
      setChecks((prev) => ({ ...prev, [n]: runs ?? [] }));
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const input = "rounded-md border border-border bg-muted px-3 py-1.5 text-sm outline-none focus:border-primary";
  const btn = "rounded-md bg-primary px-4 py-1.5 text-sm font-semibold text-background disabled:opacity-40";
  const btnSm = "rounded-md border border-border px-2 py-0.5 text-xs disabled:opacity-40";

  // Map a check conclusion/status to a color.
  const checkColor = (r: CheckRun) => {
    const s = r.Conclusion || r.Status;
    if (s === "success") return "text-green-500";
    if (s === "failure" || s === "cancelled" || s === "timed_out") return "text-red-500";
    return "text-amber-500"; // queued / in_progress / neutral
  };

  return (
    <div className="mx-auto max-w-4xl p-6 font-mono">
      <header className="mb-4 flex items-center justify-between border-b border-border pb-3">
        <h1 className="text-2xl font-bold">gitmate</h1>
        <div className="flex gap-2">
          <input value={dir} onChange={(e) => setDir(e.target.value)} placeholder="repo path (default .)" className={input} />
          <button onClick={reload} disabled={!!busy} className={btn}>Reload</button>
        </div>
      </header>

      {toast && (
        <div className={`mb-4 rounded-md px-3 py-2 text-sm ${toast.kind === "ok" ? "bg-green-600/20 text-green-400" : "bg-red-600/20 text-red-400"}`}>
          {toast.msg}
        </div>
      )}

      <Card className="mb-4">
        <CardHeader><CardTitle>Commit &amp; Push</CardTitle></CardHeader>
        <CardContent className="space-y-2">
          <div className="flex gap-2">
            <input value={commitMsg} onChange={(e) => setCommitMsg(e.target.value)} placeholder="commit message (stages all)" className={`${input} flex-1`} />
            <button onClick={doCommit} disabled={!!busy || !commitMsg.trim()} className={btn}>
              {busy === "commit" ? "…" : "Commit"}
            </button>
            <button onClick={doPush} disabled={!!busy} className={btn}>
              {busy === "push" ? "…" : "Push"}
            </button>
          </div>
        </CardContent>
      </Card>

      <Card className="mb-4">
        <CardHeader><CardTitle>New pull request</CardTitle></CardHeader>
        <CardContent className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <input value={prTitle} onChange={(e) => setPrTitle(e.target.value)} placeholder="title" className={`${input} flex-1`} />
            <input value={prHead} onChange={(e) => setPrHead(e.target.value)} placeholder="head (source branch)" className={input} />
            <input value={prBase} onChange={(e) => setPrBase(e.target.value)} placeholder="base" className={`${input} w-24`} />
            <button onClick={doPR} disabled={!!busy || !prTitle.trim() || !prHead.trim()} className={btn}>
              {busy === "pr" ? "…" : "Open PR"}
            </button>
          </div>
        </CardContent>
      </Card>

      <Card className="mb-4">
        <CardHeader><CardTitle>New issue</CardTitle></CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <input value={issueTitle} onChange={(e) => setIssueTitle(e.target.value)} placeholder="title" className={`${input} flex-1`} />
            <button onClick={doIssue} disabled={!!busy || !issueTitle.trim()} className={btn}>
              {busy === "issue" ? "…" : "Open issue"}
            </button>
          </div>
        </CardContent>
      </Card>

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
        <CardContent className="space-y-2">
          {(prs ?? []).length ? (prs ?? []).map((p) => (
            <div key={p.Number} className="border-b border-border pb-2 last:border-0">
              <div className="flex items-baseline gap-2 text-sm">
                <span className="font-semibold text-blue-500">#{p.Number}</span>
                <span className="truncate">{p.Title}</span>
                {p.Draft && <Badge variant="outline">draft</Badge>}
                <span className="text-muted-foreground">@{p.Author}</span>
                <span className="ml-auto flex gap-1">
                  <button onClick={() => loadChecks(p.Number)} disabled={!!busy} className={btnSm}>
                    {busy === `checks-${p.Number}` ? "…" : "Checks"}
                  </button>
                  <button onClick={() => doMerge(p.Number)} disabled={!!busy} className={btnSm}>
                    {busy === `merge-${p.Number}` ? "…" : "Merge"}
                  </button>
                </span>
              </div>
              {checks[p.Number] && (
                <div className="mt-1 space-y-0.5 pl-6">
                  {checks[p.Number].length ? checks[p.Number].map((r, i) => (
                    <div key={i} className="text-xs">
                      <span className={checkColor(r)}>{r.Conclusion || r.Status}</span>
                      <span className="ml-2 text-muted-foreground">{r.Name}</span>
                    </div>
                  )) : <div className="text-xs italic text-muted-foreground">no checks reported</div>}
                </div>
              )}
            </div>
          )) : <span className="italic text-muted-foreground">no open PRs</span>}
        </CardContent>
      </Card>
    </div>
  );
}