import { useEffect, useState, useCallback } from "react";
import { GitService } from "../bindings/github.com/davasorus/gitmate/gui";
import type { Status, Commit, Branch } from "../bindings/github.com/davasorus/gitmate/internal/gitops";
import type { PR, CheckRun } from "../bindings/github.com/davasorus/gitmate/internal/ghapi";

/* ---------- types ---------- */
type View = "changes" | "history" | "branches" | "prs";
type Toast = { kind: "ok" | "err"; msg: string } | null;

/* ---------- tiny shared UI atoms (extract to files later per UX.md) ---------- */

function StatusBadge({ staged, unstaged }: { staged: string; unstaged: string }) {
  const label = [staged, unstaged].filter(Boolean).join("/") || "•";
  const color =
    staged === "conflict" || unstaged === "conflict" ? "text-[var(--color-conflict)]"
    : staged === "added" ? "text-[var(--color-added)]"
    : staged === "deleted" || unstaged === "deleted" ? "text-[var(--color-removed)]"
    : "text-[var(--color-modified)]";
  return <span className={`w-20 shrink-0 text-xs ${color}`}>{label}</span>;
}

function CheckBadge({ run }: { run: CheckRun }) {
  const s = run.Conclusion || run.Status;
  const color =
    s === "success" ? "text-[var(--color-check-pass)]"
    : s === "failure" || s === "cancelled" || s === "timed_out" ? "text-[var(--color-check-fail)]"
    : "text-[var(--color-check-pending)]";
  return (
    <div className="text-xs">
      <span className={color}>{s}</span>
      <span className="ml-2 text-muted-foreground">{run.Name}</span>
    </div>
  );
}

/* ---------- app ---------- */

export default function App() {
  const [view, setView] = useState<View>("changes");
  const [dir, setDir] = useState("");

  const [status, setStatus] = useState<Status | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [prs, setPRs] = useState<PR[]>([]);
  const [checks, setChecks] = useState<Record<number, CheckRun[]>>({});

  const [toast, setToast] = useState<Toast>(null);
  const [busy, setBusy] = useState("");

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
        GitService.Log(15),
      ]);
      setStatus(s);
      setBranches(b ?? []);
      setCommits(c ?? []);
      try {
        setPRs((await GitService.PRs("open")) ?? []);
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
      const h = await GitService.Commit(commitMsg);
      setCommitMsg("");
      return `committed ${h}`;
    }, "committed");
  const doPush = () => run("push", () => GitService.Push(true), "pushed");
  const doPR = () =>
    run("pr", async () => {
      const url = await GitService.CreatePR(prTitle, "", prHead, prBase);
      setPrTitle(""); setPrHead("");
      setView("prs");
      return url;
    }, "PR opened");
  const doIssue = () =>
    run("issue", async () => {
      const url = await GitService.CreateIssue(issueTitle, "");
      setIssueTitle("");
      return url;
    }, "issue opened");
  const doMerge = (n: number) =>
    run(`merge-${n}`, async () => {
      const sha = await GitService.MergePR(n, "merge");
      return `merged #${n} (${sha.slice(0, 7)})`;
    }, `merged #${n}`);

    const loadChecks = async (n: number) => {
    setBusy(`checks-${n}`);
    try {
      const runs = (await GitService.PRChecks(n)) ?? [];
      setChecks((p) => ({ ...p, [n]: runs }));
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  /* ---------- styles ---------- */
  const input = "rounded-md border border-border bg-muted px-3 py-1.5 text-sm outline-none focus:border-[var(--color-ahead)]";
  const btn = "rounded-md bg-primary px-3 py-1.5 text-sm font-semibold text-background disabled:opacity-40";
  const btnSm = "rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40";

  const changed = (status?.Changes?.length ?? 0) + (status?.Untracked?.length ?? 0);

  /* ---------- sidebar item ---------- */
  const NavItem = ({ id, label, badge }: { id: View; label: string; badge?: number }) => (
    <button
      onClick={() => setView(id)}
      className={`flex w-full items-center justify-between px-3 py-1.5 text-left text-sm ${
        view === id ? "bg-muted font-semibold" : "hover:bg-muted/60"
      }`}
    >
      <span>{label}</span>
      {badge ? <span className="rounded bg-border px-1.5 text-xs">{badge}</span> : null}
    </button>
  );

  const Section = ({ children }: { children: string }) => (
    <div className="px-3 pt-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">{children}</div>
  );

  return (
    <div className="flex h-screen font-mono text-foreground">
      {/* ---------- SIDEBAR ---------- */}
      <aside className="flex w-60 shrink-0 flex-col border-r border-border bg-[var(--color-sidebar)]">
        <div className="border-b border-border p-3">
          <div className="text-sm font-bold">gitmate</div>
          <div className="mt-1 text-xs text-muted-foreground">
            {status?.Detached ? "detached HEAD" : status?.Branch ?? "…"}
            {status?.Upstream && (
              <span className="ml-1">
                <span className="text-[var(--color-ahead)]">↑{status.Ahead}</span>{" "}
                <span className="text-[var(--color-behind)]">↓{status.Behind}</span>
              </span>
            )}
          </div>
          <div className="mt-1 text-xs">
            {changed ? <span className="text-[var(--color-modified)]">{changed} changed</span>
                     : <span className="text-muted-foreground">clean</span>}
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto py-1">
          <NavItem id="changes" label="Changes" badge={changed} />
          <NavItem id="history" label="History" />
          <NavItem id="branches" label="Branches" badge={branches.length} />
          <NavItem id="prs" label="Pull Requests" badge={prs.length} />
          <Section>Soon</Section>
          <div className="px-3 py-1 text-sm text-muted-foreground/50">Remotes</div>
          <div className="px-3 py-1 text-sm text-muted-foreground/50">Stashes</div>
          <div className="px-3 py-1 text-sm text-muted-foreground/50">Tags</div>
        </nav>

        {/* status bar */}
        <div className="border-t border-border p-2">
          <div className="mb-1 flex gap-1">
            <input value={dir} onChange={(e) => setDir(e.target.value)} placeholder="repo path (.)"
                   className={`${input} w-full py-1 text-xs`} />
          </div>
          <div className="flex gap-1">
            <button onClick={reload} disabled={!!busy} className={`${btnSm} flex-1`}>Reload</button>
            <button onClick={doPush} disabled={!!busy} className={`${btnSm} flex-1`}>
              {busy === "push" ? "…" : "Push"}
            </button>
          </div>
        </div>
      </aside>

      {/* ---------- MAIN ---------- */}
      <main className="flex-1 overflow-y-auto">
        {toast && (
          <div className={`m-3 rounded-md px-3 py-2 text-sm ${
            toast.kind === "ok" ? "bg-[var(--color-added)]/20 text-[var(--color-added)]"
                                : "bg-[var(--color-removed)]/20 text-[var(--color-removed)]"}`}>
            {toast.msg}
          </div>
        )}

        <div className="p-4">
          {/* CHANGES (home) */}
          {view === "changes" && (
            <div className="space-y-4">
              <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Working changes</h2>
              {status ? (
                <>
                  <div className="rounded-lg border border-border">
                    {changed === 0 && <div className="p-3 text-sm italic text-muted-foreground">working tree clean</div>}
                    {(status.Changes ?? []).map((c, i) => (
                      <div key={i} className="flex items-baseline gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
                        <StatusBadge staged={c.Staged} unstaged={c.Unstaged} />
                        <span className="truncate">{c.Path}</span>
                      </div>
                    ))}
                    {(status.Untracked ?? []).map((u, i) => (
                      <div key={i} className="flex items-baseline gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
                        <span className="w-20 shrink-0 text-xs text-muted-foreground">untracked</span>
                        <span className="truncate">{u}</span>
                      </div>
                    ))}
                  </div>

                  <div className="flex gap-2">
                    <input value={commitMsg} onChange={(e) => setCommitMsg(e.target.value)}
                           placeholder="commit message (stages all)" className={`${input} flex-1`} />
                    <button onClick={doCommit} disabled={!!busy || !commitMsg.trim()} className={btn}>
                      {busy === "commit" ? "…" : "Commit"}
                    </button>
                  </div>
                </>
              ) : <div className="text-sm text-muted-foreground">…</div>}
            </div>
          )}

          {/* HISTORY */}
          {view === "history" && (
            <div className="space-y-2">
              <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">History</h2>
              <div className="rounded-lg border border-border">
                {(commits ?? []).map((c) => (
                  <div key={c.Hash} className="flex items-baseline gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
                    <span className="shrink-0 text-[var(--color-modified)]">{c.Short}</span>
                    <span className="truncate">{c.Subject}</span>
                    <span className="ml-auto shrink-0 text-xs text-muted-foreground">{c.Author}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* BRANCHES */}
          {view === "branches" && (
            <div className="space-y-2">
              <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Branches</h2>
              <div className="rounded-lg border border-border">
                {(branches ?? []).map((b) => (
                  <div key={b.Name} className="flex items-baseline gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
                    {b.IsCurrent ? <span className="text-[var(--color-added)]">●</span> : <span className="w-2" />}
                    <b>{b.Name}</b>
                    <span className="text-[var(--color-modified)]">{b.LastHash}</span>
                    {!b.Upstream && <span className="text-xs text-muted-foreground">(no upstream)</span>}
                    {b.Upstream && (b.Ahead || b.Behind) ? (
                      <span className="text-xs text-muted-foreground">↑{b.Ahead} ↓{b.Behind}</span>
                    ) : null}
                    <span className="ml-auto truncate text-xs text-muted-foreground">{b.LastSubject}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* PULL REQUESTS */}
          {view === "prs" && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Pull requests</h2>
              </div>

              {/* new PR */}
              <div className="flex flex-wrap gap-2 rounded-lg border border-border p-3">
                <input value={prTitle} onChange={(e) => setPrTitle(e.target.value)} placeholder="PR title" className={`${input} flex-1`} />
                <input value={prHead} onChange={(e) => setPrHead(e.target.value)} placeholder="head branch" className={input} />
                <input value={prBase} onChange={(e) => setPrBase(e.target.value)} placeholder="base" className={`${input} w-24`} />
                <button onClick={doPR} disabled={!!busy || !prTitle.trim() || !prHead.trim()} className={btn}>
                  {busy === "pr" ? "…" : "Open PR"}
                </button>
              </div>

              {/* new issue */}
              <div className="flex gap-2 rounded-lg border border-border p-3">
                <input value={issueTitle} onChange={(e) => setIssueTitle(e.target.value)} placeholder="issue title" className={`${input} flex-1`} />
                <button onClick={doIssue} disabled={!!busy || !issueTitle.trim()} className={btn}>
                  {busy === "issue" ? "…" : "Open issue"}
                </button>
              </div>

              {/* open PRs */}
              <div className="rounded-lg border border-border">
                {(prs ?? []).length ? (prs ?? []).map((p) => (
                  <div key={p.Number} className="border-b border-border p-3 last:border-0">
                    <div className="flex items-baseline gap-2 text-sm">
                      <span className="font-semibold text-[var(--color-ahead)]">#{p.Number}</span>
                      <span className="truncate">{p.Title}</span>
                      <span className="text-xs text-muted-foreground">@{p.Author}</span>
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
                        {checks[p.Number].length
                          ? checks[p.Number].map((r, i) => <CheckBadge key={i} run={r} />)
                          : <div className="text-xs italic text-muted-foreground">no checks reported</div>}
                      </div>
                    )}
                  </div>
                )) : <div className="p-3 text-sm italic text-muted-foreground">no open PRs</div>}
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}