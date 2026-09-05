import { useGit, cls, View } from "./context";
import { Changes } from "./views/Changes";
import { History } from "./views/History";
import { Branches } from "./views/Branches";
import { PullRequests } from "./views/PullRequests";
import { Issues } from "./views/Issues";
import { Stashes } from "./views/Stashes";
import { Conflicts } from "./views/Conflicts";
import { Tags } from "./views/Tags";

export default function App() {
  const {
    view, setView, dir, setDir, status, branches, prs, issues, stashes,
    toast, busy, run, service, reload, mergeInProgress, rebaseInProgress, conflicts,
  } = useGit();

  const changed = (status?.Changes?.length ?? 0) + (status?.Untracked?.length ?? 0);

  const doPush = () => run("push", () => service.Push(true), "pushed");
  const doFetch = () => run("fetch", () => service.Fetch(), "fetched");
  const doPull = () => run("pull", () => service.Pull(false), "pulled");
  const doMergeAbort = () => run("merge-abort", () => service.MergeAbort(), "merge aborted");
  const doRebaseContinue = () => run("rebase-continue", () => service.RebaseContinue(), "rebase continued");
  const doRebaseAbort = () => run("rebase-abort", () => service.RebaseAbort(), "rebase aborted");

  const NavItem = ({ id, label, badge }: { id: View; label: string; badge?: number }) => (
    <button onClick={() => setView(id)}
            className={`flex w-full items-center justify-between px-3 py-1.5 text-left text-sm ${view === id ? "bg-muted font-semibold" : "hover:bg-muted/60"}`}>
      <span>{label}</span>
      {badge ? <span className="rounded bg-border px-1.5 text-xs">{badge}</span> : null}
    </button>
  );

  return (
    <div className="flex h-screen font-mono text-foreground">
      {/* SIDEBAR */}
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
            {changed ? <span className="text-[var(--color-modified)]">{changed} changed</span> : <span className="text-muted-foreground">clean</span>}
          </div>
        </div>
        <nav className="flex-1 overflow-y-auto py-1">
          <NavItem id="changes" label="Changes" badge={changed} />
          {(mergeInProgress || rebaseInProgress) && <NavItem id="conflicts" label="Conflicts" badge={conflicts.length} />}
          <NavItem id="history" label="History" />
          <NavItem id="branches" label="Branches" badge={branches.length} />
          <NavItem id="prs" label="Pull Requests" badge={prs.length} />
          <NavItem id="issues" label="Issues" badge={issues.length} />
          <NavItem id="stashes" label="Stashes" badge={stashes.length} />
          <div className="px-3 pt-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Soon</div>
          <div className="px-3 py-1 text-sm text-muted-foreground/50">Remotes</div>
          <NavItem id="tags" label="Tags" />
        </nav>
        <div className="border-t border-border p-2">
          <div className="mb-1 flex gap-1">
            <input value={dir} onChange={(e) => setDir(e.target.value)} placeholder="repo path (.)" className={`${cls.input} w-full py-1 text-xs`} />
          </div>
          <div className="flex gap-1">
            <button onClick={reload} disabled={!!busy} className={`${cls.btnSm} flex-1`}>Reload</button>
            <button onClick={doFetch} disabled={!!busy} className={`${cls.btnSm} flex-1`}>{busy === "fetch" ? "…" : "Fetch"}</button>
            <button onClick={doPull} disabled={!!busy} className={`${cls.btnSm} flex-1`}>{busy === "pull" ? "…" : "Pull"}</button>
            <button onClick={doPush} disabled={!!busy} className={`${cls.btnSm} flex-1`}>{busy === "push" ? "…" : "Push"}</button>
          </div>
        </div>
      </aside>

      {/* MAIN */}
      <main className="flex-1 overflow-y-auto">
        {mergeInProgress && (
          <div className="m-3 rounded-md border border-[var(--color-conflict)] bg-[var(--color-conflict)]/10 px-3 py-2 text-sm">
            <button onClick={() => setView("conflicts")} className="font-semibold text-[var(--color-conflict)] hover:underline">Merge in progress — resolve conflicts →</button>
            {conflicts.length
              ? <div className="mt-1 text-xs text-muted-foreground">{conflicts.length} conflicted file(s): {conflicts.join(", ")}. Resolve (edit + stage in Changes), then commit — or abort.</div>
              : <div className="mt-1 text-xs text-muted-foreground">No conflicts — commit to finish the merge, or abort.</div>}
            <button onClick={doMergeAbort} disabled={!!busy} className={`mt-2 ${cls.btnSm}`}>{busy === "merge-abort" ? "…" : "Abort merge"}</button>
          </div>
        )}
        {rebaseInProgress && (
          <div className="m-3 rounded-md border border-[var(--color-conflict)] bg-[var(--color-conflict)]/10 px-3 py-2 text-sm">
            <button onClick={() => setView("conflicts")} className="font-semibold text-[var(--color-conflict)] hover:underline">Rebase in progress — resolve conflicts →</button>
            {conflicts.length
              ? <div className="mt-1 text-xs text-muted-foreground">{conflicts.length} conflicted file(s): {conflicts.join(", ")}. Resolve (Take ours/theirs), then Continue.</div>
              : <div className="mt-1 text-xs text-muted-foreground">No conflicts — Continue to replay the next commit, or Abort.</div>}
            <div className="mt-2 flex gap-2">
              <button onClick={doRebaseContinue} disabled={!!busy} className={cls.btnSm}>{busy === "rebase-continue" ? "…" : "Continue"}</button>
              <button onClick={doRebaseAbort} disabled={!!busy} className={cls.btnSm}>{busy === "rebase-abort" ? "…" : "Abort rebase"}</button>
            </div>
          </div>
        )}
        {toast && (
          <div className={`m-3 rounded-md px-3 py-2 text-sm ${toast.kind === "ok" ? "bg-[var(--color-added)]/20 text-[var(--color-added)]" : "bg-[var(--color-removed)]/20 text-[var(--color-removed)]"}`}>
            {toast.msg}
          </div>
        )}
        <div className="p-4">
          {view === "changes" && <Changes />}
          {view === "history" && <History />}
          {view === "branches" && <Branches />}
          {view === "prs" && <PullRequests />}
          {view === "issues" && <Issues />}
          {view === "stashes" && <Stashes />}
          {view === "conflicts" && <Conflicts />}
          {view === "tags" && <Tags />}
        </div>
      </main>
    </div>
  );
}