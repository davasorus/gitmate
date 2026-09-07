import { useState } from "react";
import { useGit, cls, View } from "./context";
import { Changes } from "./views/Changes";
import { History } from "./views/History";
import { Branches } from "./views/Branches";
import { PullRequests } from "./views/PullRequests";
import { Issues } from "./views/Issues";
import { Stashes } from "./views/Stashes";
import { Conflicts } from "./views/Conflicts";
import { Tags } from "./views/Tags";
import { Reflog } from "./views/Reflog";
import { Remotes } from "./views/Remotes";
import { Labels } from "./views/Labels";
import { Releases } from "./views/Releases";
import { Actions } from "./views/Actions";

// Tier-1 jumps live in the top bar; everything else is under "More".
const TIER1: { id: View; label: string }[] = [
  { id: "changes", label: "Changes" },
  { id: "prs", label: "Pull Requests" },
  { id: "branches", label: "Branches" },
  { id: "actions", label: "Actions" },
];
const MORE: { id: View; label: string }[] = [
  { id: "history", label: "History" },
  { id: "issues", label: "Issues" },
  { id: "releases", label: "Releases" },
  { id: "labels", label: "Labels" },
  { id: "stashes", label: "Stashes" },
  { id: "tags", label: "Tags" },
  { id: "reflog", label: "Reflog" },
  { id: "remotes", label: "Remotes" },
];

export default function App() {
  const {
    view,
    setView,
    dir,
    setDir,
    status,
    branches,
    prs,
    issues,
    stashes,
    tags,
    toast,
    busy,
    run,
    service,
    reload,
    mergeInProgress,
    notRepo,
    undoLabel,
    rebaseInProgress,
    cherryPickInProgress,
    revertInProgress,
    conflicts,
  } = useGit();

  const [moreOpen, setMoreOpen] = useState(false);
  const [recent, setRecent] = useState<string[]>([]);
  const [recentOpen, setRecentOpen] = useState(false);

  // native folder picker → point the app at a repo, remember it
  const openRepo = async () => {
    const path = await service.SelectDirectory();
    if (!path) return;
    setDir(path);
    await service.AddRecentRepo(path);
    setRecentOpen(false);
  };
  const pickRecent = async (path: string) => {
    setDir(path);
    await service.AddRecentRepo(path);
    setRecentOpen(false);
  };
  const loadRecent = async () => {
    setRecent((await service.RecentRepos()) ?? []);
    setRecentOpen((v) => !v);
  };
  const [showDir, setShowDir] = useState(false);

  const changed = (status?.Changes?.length ?? 0) + (status?.Untracked?.length ?? 0);
  const inProgress =
    mergeInProgress || rebaseInProgress || cherryPickInProgress || revertInProgress;

  const doPush = () => run("push", () => service.Push(true), "pushed");
  const doFetch = () => run("fetch", () => service.Fetch(), "fetched");
  const doPull = () => run("pull", () => service.Pull(false), "pulled");
  const doMergeAbort = () => run("merge-abort", () => service.MergeAbort(), "merge aborted");
  const doUndo = () => {
    if (
      !window.confirm(
        `Undo "${undoLabel}"? This resets to before that operation (recoverable via reflog).`,
      )
    )
      return;
    run("undo", () => service.Undo(), "undone");
  };
  const doRebaseContinue = () =>
    run("rebase-continue", () => service.RebaseContinue(), "rebase continued");
  const doRebaseAbort = () => run("rebase-abort", () => service.RebaseAbort(), "rebase aborted");
  const doCherryContinue = () =>
    run("cp-continue", () => service.CherryPickContinue(), "cherry-pick continued");
  const doCherryAbort = () =>
    run("cp-abort", () => service.CherryPickAbort(), "cherry-pick aborted");
  const doRevertContinue = () =>
    run("rv-continue", () => service.RevertContinue(), "revert continued");
  const doRevertAbort = () => run("rv-abort", () => service.RevertAbort(), "revert aborted");

  const badgeFor = (id: View): number | undefined => {
    switch (id) {
      case "changes":
        return changed || undefined;
      case "prs":
        return prs.length || undefined;
      case "branches":
        return branches.length || undefined;
      case "issues":
        return issues.length || undefined;
      case "stashes":
        return stashes.length || undefined;
      case "tags":
        return tags.length || undefined;
      default:
        return undefined;
    }
  };

  const Jump = ({ id, label }: { id: View; label: string }) => {
    const badge = badgeFor(id);
    const active = view === id;
    return (
      <button
        onClick={() => {
          setView(id);
          setMoreOpen(false);
        }}
        className={`flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-[13px] font-medium transition-colors ${
          active
            ? "bg-[var(--color-accent-dim)] text-[var(--color-foreground)]"
            : "text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] hover:text-[var(--color-foreground)]"
        }`}
      >
        {label}
        {badge != null && (
          <span className="font-mono text-[10.5px] text-[var(--color-faint)]">{badge}</span>
        )}
      </button>
    );
  };

  return (
    <div className="flex h-screen flex-col text-[var(--color-foreground)]">
      {/* ---------------- TOP BAR ---------------- */}
      <header className="flex items-center gap-3.5 border-b border-border px-4 py-2.5">
        <div className="grid h-[22px] w-[22px] place-items-center rounded-md bg-gradient-to-br from-[var(--color-accent)] to-[#5a6ad0] text-[12px] font-semibold text-white">
          g
        </div>
        <button
          onClick={() => setShowDir((v) => !v)}
          className="flex items-center gap-1.5 text-[13.5px] font-medium hover:opacity-80"
          title="change repo"
        >
          <span className="text-[var(--color-accent)]">⑂</span>
          {status?.Detached ? "detached HEAD" : (status?.Branch ?? "…")}
        </button>
        {status?.Upstream && (
          <span className="font-mono text-[11.5px] text-[var(--color-faint)]">
            <span className="text-[var(--color-added)]">↑{status.Ahead}</span> ↓{status.Behind}
          </span>
        )}
        {showDir && (
          <input
            autoFocus
            value={dir}
            onChange={(e) => setDir(e.target.value)}
            onBlur={() => setShowDir(false)}
            placeholder="repo path (.)"
            className={`${cls.input} w-64 py-1 text-xs`}
          />
        )}

        <div className="relative">
          <button
            onClick={openRepo}
            className="rounded-lg border border-border px-2.5 py-1.5 text-[12px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)]"
            title="open a repository folder"
          >
            Open…
          </button>
        </div>
        <div className="relative">
          <button
            onClick={loadRecent}
            className="rounded-lg px-1.5 py-1.5 text-[12px] text-[var(--color-faint)] hover:bg-[var(--color-muted)]"
            title="recent repositories"
          >
            ▾
          </button>
          {recentOpen && (
            <div className="absolute left-0 top-full z-20 mt-1 w-80 rounded-lg border border-border bg-[var(--color-background)] p-1 shadow-xl">
              {recent.length === 0 ? (
                <div className="px-2.5 py-2 text-[12px] text-[var(--color-faint)]">
                  no recent repositories
                </div>
              ) : (
                recent.map((rp) => (
                  <button
                    key={rp}
                    onClick={() => pickRecent(rp)}
                    className="block w-full truncate rounded-md px-2.5 py-1.5 text-left font-mono text-[11.5px] text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] hover:text-[var(--color-foreground)]"
                    title={rp}
                  >
                    {rp}
                  </button>
                ))
              )}
            </div>
          )}
        </div>

        {/* jumps */}
        <nav className="ml-2 flex items-center gap-0.5">
          {TIER1.map((t) => (
            <Jump key={t.id} {...t} />
          ))}
          {inProgress && <Jump id="conflicts" label="Conflicts" />}
          <div className="relative">
            <button
              onClick={() => setMoreOpen((v) => !v)}
              className={`rounded-lg px-2.5 py-1.5 text-[13px] font-medium transition-colors ${
                MORE.some((m) => m.id === view)
                  ? "bg-[var(--color-accent-dim)] text-[var(--color-foreground)]"
                  : "text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] hover:text-[var(--color-foreground)]"
              }`}
            >
              More…
            </button>
            {moreOpen && (
              <div className="absolute left-0 top-full z-20 mt-1 w-44 rounded-lg border border-border bg-[var(--color-background)] p-1 shadow-xl">
                {MORE.map((m) => (
                  <button
                    key={m.id}
                    onClick={() => {
                      setView(m.id);
                      setMoreOpen(false);
                    }}
                    className={`flex w-full items-center justify-between rounded-md px-2.5 py-1.5 text-left text-[13px] ${
                      view === m.id
                        ? "bg-[var(--color-accent-dim)] text-[var(--color-foreground)]"
                        : "text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] hover:text-[var(--color-foreground)]"
                    }`}
                  >
                    {m.label}
                    {badgeFor(m.id) != null && (
                      <span className="font-mono text-[10.5px] text-[var(--color-faint)]">
                        {badgeFor(m.id)}
                      </span>
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>
        </nav>

        <div className="flex-1" />

        {/* right: undo + sync */}
        <div className="flex items-center gap-1.5">
          {undoLabel && (
            <button
              onClick={doUndo}
              disabled={!!busy}
              title={`Undo: ${undoLabel}`}
              className="rounded-lg border border-border px-2.5 py-1.5 text-[12px] font-medium text-[var(--color-modified)] hover:bg-[var(--color-muted)] disabled:opacity-40"
            >
              {busy === "undo" ? "…" : "↩ Undo"}
            </button>
          )}
          <button
            onClick={reload}
            disabled={!!busy}
            className="rounded-lg border border-border px-2.5 py-1.5 text-[12px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
          >
            Reload
          </button>
          <button
            onClick={doFetch}
            disabled={!!busy}
            className="rounded-lg border border-border px-2.5 py-1.5 text-[12px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
          >
            {busy === "fetch" ? "…" : "Fetch"}
          </button>
          <button
            onClick={doPull}
            disabled={!!busy}
            className="rounded-lg border border-border px-2.5 py-1.5 text-[12px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
          >
            {busy === "pull" ? "…" : "Pull"}
          </button>
          <button
            onClick={doPush}
            disabled={!!busy}
            className="rounded-lg bg-[var(--color-accent)] px-2.5 py-1.5 text-[12px] font-semibold text-white hover:opacity-90 disabled:opacity-40"
          >
            {busy === "push" ? "…" : `Push${status?.Ahead ? " ↑" + status.Ahead : ""}`}
          </button>
        </div>
      </header>

      {/* ---------------- MAIN ---------------- */}
      <main className="min-h-0 flex-1 overflow-y-auto">
        {notRepo && (
          <div className="m-3 rounded-lg border border-border bg-[var(--color-card)] px-4 py-3 text-sm">
            <div className="font-semibold text-[var(--color-modified)]">Not a git repository</div>
            <div className="mt-1 text-[var(--color-muted-foreground)]">
              The folder <span className="font-mono">{dir || "."}</span> isn't a git repo.
            </div>
            <button
              onClick={openRepo}
              className="mt-2 rounded-lg bg-[var(--color-accent)] px-3 py-1.5 text-[12.5px] font-semibold text-white hover:opacity-90"
            >
              Open repository…
            </button>
          </div>
        )}
        {/* in-progress banners */}
        {mergeInProgress && (
          <div className="m-3 rounded-lg border border-[var(--color-conflict)] bg-[var(--color-conflict)]/10 px-3 py-2 text-sm">
            <button
              onClick={() => setView("conflicts")}
              className="font-semibold text-[var(--color-conflict)] hover:underline"
            >
              Merge in progress — resolve conflicts →
            </button>
            {conflicts.length ? (
              <div className="mt-1 text-xs text-muted-foreground">
                {conflicts.length} conflicted file(s): {conflicts.join(", ")}. Resolve, then commit
                — or abort.
              </div>
            ) : (
              <div className="mt-1 text-xs text-muted-foreground">
                No conflicts — commit to finish the merge, or abort.
              </div>
            )}
            <button onClick={doMergeAbort} disabled={!!busy} className={`mt-2 ${cls.btnSm}`}>
              {busy === "merge-abort" ? "…" : "Abort merge"}
            </button>
          </div>
        )}
        {rebaseInProgress && (
          <div className="m-3 rounded-lg border border-[var(--color-conflict)] bg-[var(--color-conflict)]/10 px-3 py-2 text-sm">
            <button
              onClick={() => setView("conflicts")}
              className="font-semibold text-[var(--color-conflict)] hover:underline"
            >
              Rebase in progress — resolve conflicts →
            </button>
            <div className="mt-2 flex gap-2">
              <button onClick={doRebaseContinue} disabled={!!busy} className={cls.btnSm}>
                {busy === "rebase-continue" ? "…" : "Continue"}
              </button>
              <button onClick={doRebaseAbort} disabled={!!busy} className={cls.btnSm}>
                {busy === "rebase-abort" ? "…" : "Abort"}
              </button>
            </div>
          </div>
        )}
        {cherryPickInProgress && (
          <div className="m-3 rounded-lg border border-[var(--color-conflict)] bg-[var(--color-conflict)]/10 px-3 py-2 text-sm">
            <span className="font-semibold text-[var(--color-conflict)]">
              Cherry-pick in progress
            </span>
            <div className="mt-2 flex gap-2">
              <button onClick={doCherryContinue} disabled={!!busy} className={cls.btnSm}>
                {busy === "cp-continue" ? "…" : "Continue"}
              </button>
              <button onClick={doCherryAbort} disabled={!!busy} className={cls.btnSm}>
                {busy === "cp-abort" ? "…" : "Abort"}
              </button>
            </div>
          </div>
        )}
        {revertInProgress && (
          <div className="m-3 rounded-lg border border-[var(--color-conflict)] bg-[var(--color-conflict)]/10 px-3 py-2 text-sm">
            <span className="font-semibold text-[var(--color-conflict)]">Revert in progress</span>
            <div className="mt-2 flex gap-2">
              <button onClick={doRevertContinue} disabled={!!busy} className={cls.btnSm}>
                {busy === "rv-continue" ? "…" : "Continue"}
              </button>
              <button onClick={doRevertAbort} disabled={!!busy} className={cls.btnSm}>
                {busy === "rv-abort" ? "…" : "Abort"}
              </button>
            </div>
          </div>
        )}

        {toast && (
          <div
            className={`m-3 rounded-lg px-3 py-2 text-sm ${toast.kind === "ok" ? "bg-[var(--color-added)]/20 text-[var(--color-added)]" : "bg-[var(--color-removed)]/20 text-[var(--color-removed)]"}`}
          >
            {toast.msg}
          </div>
        )}

        <div className="p-4">
          {!notRepo && view === "changes" && <Changes />}
          {!notRepo && view === "history" && <History />}
          {!notRepo && view === "branches" && <Branches />}
          {!notRepo && view === "prs" && <PullRequests />}
          {!notRepo && view === "issues" && <Issues />}
          {!notRepo && view === "stashes" && <Stashes />}
          {!notRepo && view === "conflicts" && <Conflicts />}
          {!notRepo && view === "tags" && <Tags />}
          {!notRepo && view === "remotes" && <Remotes />}
          {!notRepo && view === "labels" && <Labels />}
          {!notRepo && view === "releases" && <Releases />}
          {!notRepo && view === "actions" && <Actions />}
          {!notRepo && view === "reflog" && <Reflog />}
        </div>
      </main>
    </div>
  );
}
