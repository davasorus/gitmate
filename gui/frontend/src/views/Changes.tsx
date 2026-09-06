import React, { useState } from "react";
import { useGit, cls } from "../context";
import { StatusBadge } from "../components/StatusBadge";
import { DiffView } from "../components/DiffView";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { FileDiff } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function Changes() {
  const { status, busy, run, service, flash, setBusy } = useGit();
  const [openDiff, setOpenDiff] = useState<{ path: string; staged: boolean } | null>(null);
  const [diffFiles, setDiffFiles] = useState<FileDiff[]>([]);
  const [confirmDiscard, setConfirmDiscard] = useState<string | null>(null);
  const [commitMsg, setCommitMsg] = useState("");

  const showDiff = async (path: string, staged: boolean) => {
    if (openDiff && openDiff.path === path && openDiff.staged === staged) {
      setOpenDiff(null);
      setDiffFiles([]);
      return;
    }
    setBusy(`diff-${path}`);
    try {
      setDiffFiles((await service.Diff(path, staged)) ?? []);
      setOpenDiff({ path, staged });
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const doStage = (p: string) => run(`stage-${p}`, () => service.StagePath(p), `staged ${p}`);
  const doUnstage = (p: string) =>
    run(`unstage-${p}`, () => service.UnstagePath(p), `unstaged ${p}`);
  const doDiscard = (p: string) =>
    run(
      `discard-${p}`,
      async () => {
        await service.DiscardPath(p);
        setConfirmDiscard(null);
        return `discarded ${p}`;
      },
      `discarded ${p}`,
    );
  const doCommit = () =>
    run(
      "commit",
      async () => {
        await service.Stage();
        const h = await service.Commit(commitMsg);
        setCommitMsg("");
        return `committed ${h}`;
      },
      "committed",
    );

  if (!status) return <div className="text-sm text-muted-foreground">…</div>;

  const chs = status.Changes ?? [];
  const staged = chs.filter((c) => (c.Staged ?? "") !== "");
  const unstaged = chs.filter((c) => (c.Unstaged ?? "") !== "");
  const untracked = status.Untracked ?? [];

  const Row = ({
    path,
    isStaged,
    badge,
  }: {
    path: string;
    isStaged: boolean;
    badge: React.ReactNode;
  }) => (
    <div>
      <div className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0 hover:bg-muted/60">
        <button
          onClick={() => (isStaged ? doUnstage(path) : doStage(path))}
          disabled={!!busy}
          className="w-5 shrink-0 rounded border border-border text-center text-xs hover:bg-muted disabled:opacity-40"
          title={isStaged ? "unstage" : "stage"}
        >
          {busy === `${isStaged ? "unstage" : "stage"}-${path}` ? "…" : isStaged ? "\u2212" : "+"}
        </button>
        {!isStaged && (
          <button
            onClick={() => setConfirmDiscard(path)}
            disabled={!!busy}
            className="w-5 shrink-0 rounded border border-border text-center text-xs text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10 disabled:opacity-40"
            title="discard changes"
          >
            {busy === `discard-${path}` ? "…" : "\u2717"}
          </button>
        )}
        <button
          onClick={() => showDiff(path, isStaged)}
          className="flex flex-1 items-baseline gap-2 text-left"
        >
          {badge}
          <span className="truncate">{path}</span>
          {busy === `diff-${path}` && (
            <span className="ml-auto text-xs text-muted-foreground">…</span>
          )}
        </button>
      </div>
      {openDiff?.path === path && openDiff.staged === isStaged && (
        <div className="border-b border-border bg-background px-2 py-2">
          <DiffView files={diffFiles} />
        </div>
      )}
    </div>
  );

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
        Working changes
      </h2>

      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Staged ({staged.length})
        </div>
        <div className="rounded-lg border border-border">
          {staged.length === 0 ? (
            <div className="p-3 text-xs italic text-muted-foreground">nothing staged</div>
          ) : (
            staged.map((c) => (
              <Row
                key={`s-${c.Path}`}
                path={c.Path}
                isStaged
                badge={<StatusBadge staged={c.Staged} unstaged="" />}
              />
            ))
          )}
        </div>
      </div>

      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Unstaged ({unstaged.length + untracked.length})
        </div>
        <div className="rounded-lg border border-border">
          {unstaged.length === 0 && untracked.length === 0 ? (
            <div className="p-3 text-xs italic text-muted-foreground">nothing unstaged</div>
          ) : (
            <>
              {unstaged.map((c) => (
                <Row
                  key={`u-${c.Path}`}
                  path={c.Path}
                  isStaged={false}
                  badge={<StatusBadge staged="" unstaged={c.Unstaged} />}
                />
              ))}
              {untracked.map((u) => (
                <div
                  key={`t-${u}`}
                  className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0 hover:bg-muted/60"
                >
                  <button
                    onClick={() => doStage(u)}
                    disabled={!!busy}
                    className="w-5 shrink-0 rounded border border-border text-center text-xs hover:bg-muted disabled:opacity-40"
                    title="stage"
                  >
                    {busy === `stage-${u}` ? "…" : "+"}
                  </button>
                  <span className="w-20 shrink-0 text-xs text-muted-foreground">untracked</span>
                  <span className="truncate">{u}</span>
                </div>
              ))}
            </>
          )}
        </div>
      </div>

      <div className="flex gap-2">
        <input
          value={commitMsg}
          onChange={(e) => setCommitMsg(e.target.value)}
          placeholder="commit message (stages all)"
          className={`${cls.input} flex-1`}
        />
        <button onClick={doCommit} disabled={!!busy || !commitMsg.trim()} className={cls.btn}>
          {busy === "commit" ? "…" : "Commit"}
        </button>
      </div>

      {confirmDiscard && (
        <ConfirmDialog
          title="Discard changes?"
          body={
            <>
              This permanently deletes your uncommitted changes to{" "}
              <span className="text-foreground">{confirmDiscard}</span>. This cannot be undone.
            </>
          }
          confirmLabel="Discard"
          busy={!!busy}
          onCancel={() => setConfirmDiscard(null)}
          onConfirm={() => doDiscard(confirmDiscard)}
        />
      )}
    </div>
  );
}
