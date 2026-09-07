import { useState } from "react";
import { useGit } from "../context";
import { DiffView } from "../components/DiffView";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { FileDiff, Hunk } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function Changes() {
  const { status, busy, run, service, flash, setBusy } = useGit();
  const [sel, setSel] = useState<{ path: string; staged: boolean } | null>(null);
  const [diffFiles, setDiffFiles] = useState<FileDiff[]>([]);
  const [confirmDiscard, setConfirmDiscard] = useState<string | null>(null);
  const [commitMsg, setCommitMsg] = useState("");

  const openDiff = async (path: string, staged: boolean) => {
    setSel({ path, staged });
    setBusy(`diff-${path}`);
    try {
      setDiffFiles((await service.Diff(path, staged)) ?? []);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const refreshDiff = async () => {
    if (sel) setDiffFiles((await service.Diff(sel.path, sel.staged)) ?? []);
  };

  const doStage = (p: string) => run(`stage-${p}`, () => service.StagePath(p), `staged ${p}`);
  const doUnstage = (p: string) =>
    run(`unstage-${p}`, () => service.UnstagePath(p), `unstaged ${p}`);
  const doStageHunk = (path: string, hunk: Hunk) =>
    run(
      `stagehunk-${path}`,
      async () => {
        await service.StageHunk(path, hunk);
        await refreshDiff();
        return "staged hunk";
      },
      "staged hunk",
    );
  const doUnstageHunk = (path: string, hunk: Hunk) =>
    run(
      `unstagehunk-${path}`,
      async () => {
        await service.UnstageHunk(path, hunk);
        await refreshDiff();
        return "unstaged hunk";
      },
      "unstaged hunk",
    );
  const doDiscard = (p: string) =>
    run(
      `discard-${p}`,
      async () => {
        await service.DiscardPath(p);
        setConfirmDiscard(null);
        if (sel?.path === p) setSel(null);
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
  const stagedCount = staged.length;

  const statusGlyph = (code: string) =>
    code === "A" ? (
      <span className="font-mono text-[11px] text-[var(--color-added)]">A</span>
    ) : code === "D" ? (
      <span className="font-mono text-[11px] text-[var(--color-removed)]">D</span>
    ) : (
      <span className="font-mono text-[11px] text-[var(--color-modified)]">M</span>
    );

  // a single file row
  const FileRow = ({
    path,
    code,
    isStaged,
    untrackedFlag,
  }: {
    path: string;
    code: string;
    isStaged: boolean;
    untrackedFlag?: boolean;
  }) => {
    const selected = sel?.path === path && sel?.staged === isStaged;
    const parts = path.split("/");
    const name = parts.pop();
    const dir = parts.length ? parts.join("/") + "/" : "";
    return (
      <div
        onClick={() => openDiff(path, isStaged)}
        className={`group flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-1.5 ${
          selected ? "bg-[var(--color-accent-dim)]" : "hover:bg-[var(--color-card)]"
        }`}
      >
        {untrackedFlag ? (
          <span className="font-mono text-[11px] text-[var(--color-added)]">＋</span>
        ) : (
          statusGlyph(code)
        )}
        <span className="min-w-0 flex-1 truncate text-[12.5px]">
          <span className="text-[var(--color-faint)]">{dir}</span>
          <span>{name}</span>
        </span>
        <span className="flex shrink-0 items-center gap-1 opacity-0 group-hover:opacity-100">
          {isStaged ? (
            <button
              onClick={(e) => {
                e.stopPropagation();
                doUnstage(path);
              }}
              disabled={!!busy}
              className="rounded px-1.5 py-0.5 text-[11px] text-[var(--color-accent)] hover:bg-[var(--color-muted)]"
            >
              unstage
            </button>
          ) : (
            <>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  doStage(path);
                }}
                disabled={!!busy}
                className="rounded px-1.5 py-0.5 text-[11px] text-[var(--color-accent)] hover:bg-[var(--color-muted)]"
              >
                stage
              </button>
              {!untrackedFlag && (
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setConfirmDiscard(path);
                  }}
                  disabled={!!busy}
                  className="rounded px-1.5 py-0.5 text-[11px] text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10"
                >
                  discard
                </button>
              )}
            </>
          )}
        </span>
      </div>
    );
  };

  return (
    <div className="flex h-[calc(100vh-116px)] gap-0">
      {/* LEFT: file list + pinned commit box */}
      <div className="flex w-[320px] shrink-0 flex-col border-r border-border pr-3">
        <div className="min-h-0 flex-1 overflow-y-auto">
          {/* staged */}
          <div className="mb-1 flex items-center gap-2 px-2.5 pt-1">
            <span className="text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
              STAGED
            </span>
            <span className="font-mono text-[11px] text-[var(--color-faint)]">{staged.length}</span>
            {staged.length > 0 && (
              <button
                onClick={() => staged.forEach((c) => doUnstage(c.Path))}
                className="ml-auto text-[11px] font-medium text-[var(--color-accent)] hover:underline"
              >
                Unstage all
              </button>
            )}
          </div>
          {staged.length === 0 ? (
            <div className="px-2.5 py-1 text-[11.5px] italic text-[var(--color-faint)]">
              nothing staged
            </div>
          ) : (
            staged.map((c) => (
              <FileRow key={`s-${c.Path}`} path={c.Path} code={c.Staged} isStaged />
            ))
          )}

          {/* unstaged */}
          <div className="mb-1 mt-3 flex items-center gap-2 px-2.5">
            <span className="text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
              CHANGES
            </span>
            <span className="font-mono text-[11px] text-[var(--color-faint)]">
              {unstaged.length + untracked.length}
            </span>
            {unstaged.length + untracked.length > 0 && (
              <button
                onClick={() => run("stage-all", () => service.Stage(), "staged all")}
                className="ml-auto text-[11px] font-medium text-[var(--color-accent)] hover:underline"
              >
                Stage all
              </button>
            )}
          </div>
          {unstaged.length === 0 && untracked.length === 0 ? (
            <div className="px-2.5 py-1 text-[11.5px] italic text-[var(--color-faint)]">
              nothing to stage
            </div>
          ) : (
            <>
              {unstaged.map((c) => (
                <FileRow key={`u-${c.Path}`} path={c.Path} code={c.Unstaged} isStaged={false} />
              ))}
              {untracked.map((u) => (
                <FileRow key={`t-${u}`} path={u} code="A" isStaged={false} untrackedFlag />
              ))}
            </>
          )}
        </div>

        {/* commit box, pinned */}
        <div className="mt-3 flex flex-col gap-2 border-t border-border pt-3">
          <textarea
            value={commitMsg}
            onChange={(e) => setCommitMsg(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && (e.ctrlKey || e.metaKey) && commitMsg.trim() && !busy) {
                e.preventDefault();
                doCommit();
              }
            }}
            placeholder={"Commit message…  (⌘↵ to commit)"}
            rows={3}
            className="resize-none rounded-lg border border-border bg-[var(--color-card)] px-3 py-2 text-[13px] outline-none focus:border-[var(--color-accent)] focus:bg-[var(--color-muted)]"
          />
          <button
            onClick={doCommit}
            disabled={!!busy || !commitMsg.trim()}
            className="rounded-lg bg-[var(--color-accent)] px-3 py-2 text-[13px] font-semibold text-white hover:opacity-90 disabled:opacity-40"
          >
            {busy === "commit"
              ? "…"
              : stagedCount > 0
                ? `Commit ${stagedCount} staged`
                : "Commit all changes"}
          </button>
        </div>
      </div>

      {/* RIGHT: diff reader */}
      <div className="min-w-0 flex-1 overflow-y-auto pl-4">
        {!sel ? (
          <div className="grid h-full place-items-center text-[13px] text-[var(--color-faint)]">
            Select a file to view its diff
          </div>
        ) : busy === `diff-${sel.path}` ? (
          <div className="grid h-full place-items-center text-[13px] text-[var(--color-faint)]">
            …
          </div>
        ) : (
          <DiffView
            files={diffFiles}
            hunkAction={
              sel.staged
                ? { label: "Unstage hunk", onClick: doUnstageHunk, disabled: !!busy }
                : { label: "Stage hunk", onClick: doStageHunk, disabled: !!busy }
            }
          />
        )}
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
