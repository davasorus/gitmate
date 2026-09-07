import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Commit } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function Branches() {
  const { branches, busy, run, service, flash, setBusy } = useGit();
  const [newBranch, setNewBranch] = useState("");
  const [showNew, setShowNew] = useState(false);
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  const [renaming, setRenaming] = useState<{ old: string; next: string } | null>(null);
  const [sel, setSel] = useState<string | null>(null);
  const [recent, setRecent] = useState<Commit[]>([]);

  // fetch recent commits for the selected branch (remote-only branches resolve
  // via their origin/<name> ref)
  useEffect(() => {
    if (!sel) {
      setRecent([]);
      return;
    }
    const b = (branches ?? []).find((x) => x.Name === sel);
    const ref = b && b.IsRemote && !b.IsLocal ? `${b.Remote}/${b.Name}` : sel;
    let cancelled = false;
    (async () => {
      setBusy(`branch-log-${sel}`);
      try {
        const cs = (await service.LogRef(ref, 8)) ?? [];
        if (!cancelled) setRecent(cs);
      } catch (e) {
        if (!cancelled) {
          setRecent([]);
          flash("err", String(e));
        }
      } finally {
        if (!cancelled) setBusy("");
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sel]);

  const doSwitch = (b: string) => run(`switch-${b}`, () => service.Switch(b), `switched to ${b}`);
  const doCreate = () =>
    run(
      "create-branch",
      async () => {
        const b = newBranch.trim();
        await service.SwitchNew(b);
        setNewBranch("");
        setShowNew(false);
        return `created ${b}`;
      },
      "created",
    );
  const doDelete = (name: string, force: boolean) =>
    run(
      `delbranch-${name}`,
      async () => {
        await service.DeleteBranch(name, force);
        setConfirmDel(null);
        if (sel === name) setSel(null);
        return `deleted ${name}`;
      },
      "deleted",
    );
  const doRename = () =>
    run(
      "rename-branch",
      async () => {
        const { old, next } = renaming!;
        await service.RenameBranch(old, next.trim());
        setRenaming(null);
        return `renamed to ${next.trim()}`;
      },
      "renamed",
    );
  const doMerge = (b: string) =>
    run(
      `merge-branch-${b}`,
      async () => {
        await service.Merge(b);
        const c = (await service.ConflictedFiles()) ?? [];
        return c.length ? `merge started — ${c.length} conflict(s)` : `merged ${b}`;
      },
      "merged",
    );
  const doRebase = (b: string) =>
    run(
      `rebase-branch-${b}`,
      async () => {
        await service.Rebase(b);
        const c = (await service.ConflictedFiles()) ?? [];
        return c.length ? `rebase started — ${c.length} conflict(s)` : `rebased onto ${b}`;
      },
      "rebased",
    );

  const list = branches ?? [];
  const local = list.filter((b) => b.IsLocal);
  const remoteOnly = list.filter((b) => b.IsRemote && !b.IsLocal);
  const selected = list.find((b) => b.Name === sel) ?? null;

  const Row = ({ b }: { b: (typeof list)[number] }) => {
    const active = sel === b.Name;
    return (
      <div
        onClick={() => setSel(b.Name)}
        className={`flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 ${
          active ? "bg-[var(--color-accent-dim)]" : "hover:bg-[var(--color-card)]"
        }`}
      >
        <span className={b.IsCurrent ? "text-[var(--color-added)]" : "text-[var(--color-faint)]"}>
          {b.IsCurrent ? "●" : "⑂"}
        </span>
        <span className="min-w-0 flex-1 truncate text-[12.5px] font-medium">{b.Name}</span>
        {b.IsCurrent && (
          <span className="rounded-full bg-[var(--color-accent-dim)] px-1.5 py-0.5 text-[10px] text-[var(--color-accent)]">
            current
          </span>
        )}
        {b.IsRemote && !b.IsLocal && (
          <span className="rounded-full border border-border px-1.5 py-0.5 text-[10px] text-[var(--color-faint)]">
            remote
          </span>
        )}
        {b.Upstream && (b.Ahead || b.Behind) ? (
          <span className="font-mono text-[10.5px] text-[var(--color-faint)]">
            ↑{b.Ahead} ↓{b.Behind}
          </span>
        ) : null}
      </div>
    );
  };

  return (
    <div className="flex h-[calc(100vh-116px)] gap-0">
      {/* LEFT: branch list */}
      <div className="flex w-[320px] shrink-0 flex-col border-r border-border pr-3">
        <div className="mb-2 flex items-center px-1">
          <span className="text-[13px] font-semibold tracking-[-0.01em]">Branches</span>
          <button
            onClick={() => setShowNew((v) => !v)}
            className="ml-auto text-[11.5px] font-semibold text-[var(--color-accent)] hover:underline"
          >
            {showNew ? "Cancel" : "+ New branch"}
          </button>
        </div>
        {showNew && (
          <div className="mb-2 flex gap-2 px-1">
            <input
              autoFocus
              value={newBranch}
              onChange={(e) => setNewBranch(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && newBranch.trim()) doCreate();
              }}
              placeholder="new branch name"
              className={`${cls.input} flex-1 py-1 text-xs`}
            />
            <button
              onClick={doCreate}
              disabled={!!busy || !newBranch.trim()}
              className="rounded-lg bg-[var(--color-accent)] px-2.5 py-1 text-[12px] font-semibold text-white disabled:opacity-40"
            >
              {busy === "create-branch" ? "…" : "Create"}
            </button>
          </div>
        )}

        <div className="min-h-0 flex-1 overflow-y-auto">
          <div className="px-2.5 pb-1 pt-1 text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
            LOCAL
          </div>
          {local.map((b) => (
            <Row key={`l-${b.Name}`} b={b} />
          ))}
          {remoteOnly.length > 0 && (
            <>
              <div className="px-2.5 pb-1 pt-3 text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
                REMOTE ONLY
              </div>
              {remoteOnly.map((b) => (
                <Row key={`r-${b.Name}`} b={b} />
              ))}
            </>
          )}
        </div>
      </div>

      {/* RIGHT: branch detail */}
      <div className="min-w-0 flex-1 overflow-y-auto pl-4">
        {!selected ? (
          <div className="grid h-full place-items-center text-[13px] text-[var(--color-faint)]">
            Select a branch
          </div>
        ) : (
          <div className="space-y-4">
            <div className="border-b border-border pb-3">
              <div className="flex items-center gap-2 text-[16px] font-semibold tracking-[-0.02em]">
                <span className="text-[var(--color-accent)]">⑂</span>
                {selected.Name}
                {selected.IsCurrent && (
                  <span className="rounded-full bg-[var(--color-accent-dim)] px-2 py-0.5 text-[11px] text-[var(--color-accent)]">
                    current
                  </span>
                )}
              </div>
              <div className="mt-1.5 font-mono text-[12px] text-[var(--color-muted-foreground)]">
                {selected.Upstream ? (
                  <>
                    tracks{" "}
                    <span className="text-[var(--color-foreground)]">{selected.Upstream}</span> ·{" "}
                    <span className="text-[var(--color-added)]">↑{selected.Ahead}</span> ↓
                    {selected.Behind}
                  </>
                ) : selected.IsRemote && !selected.IsLocal ? (
                  <>only on {selected.Remote} — Switch creates a local tracking branch</>
                ) : (
                  <>no upstream</>
                )}
              </div>
            </div>

            <div>
              <div className="mb-2 text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
                RECENT COMMITS
              </div>
              {busy === `branch-log-${selected.Name}` ? (
                <div className="text-[12px] text-[var(--color-faint)]">…</div>
              ) : recent.length === 0 ? (
                <div className="text-[12px] italic text-[var(--color-faint)]">no commits</div>
              ) : (
                recent.map((c) => (
                  <div
                    key={c.Hash}
                    className="flex items-center gap-3 border-b border-border py-2 last:border-0"
                  >
                    <span className="grid h-[18px] w-[18px] shrink-0 place-items-center rounded-full bg-[var(--color-raised)] text-[9px] text-[var(--color-muted-foreground)]">
                      {c.Author.slice(0, 2).toUpperCase()}
                    </span>
                    <span className="font-mono text-[11px] text-[var(--color-modified)]">
                      {c.Short}
                    </span>
                    <span className="min-w-0 flex-1 truncate text-[12.5px]">{c.Subject}</span>
                  </div>
                ))
              )}
            </div>

            <div className="flex flex-wrap gap-2">
              {!selected.IsCurrent && (
                <button
                  onClick={() => doSwitch(selected.Name)}
                  disabled={!!busy}
                  className="rounded-lg bg-[var(--color-accent)] px-3.5 py-1.5 text-[12.5px] font-semibold text-white hover:opacity-90 disabled:opacity-40"
                >
                  {busy === `switch-${selected.Name}`
                    ? "…"
                    : selected.IsRemote && !selected.IsLocal
                      ? "Checkout"
                      : "Switch"}
                </button>
              )}
              {!selected.IsCurrent && (
                <button
                  onClick={() => doMerge(selected.Name)}
                  disabled={!!busy}
                  className="rounded-lg border border-border px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
                >
                  {busy === `merge-branch-${selected.Name}` ? "…" : "Merge into current"}
                </button>
              )}
              {!selected.IsCurrent && (
                <button
                  onClick={() => doRebase(selected.Name)}
                  disabled={!!busy}
                  className="rounded-lg border border-border px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
                >
                  {busy === `rebase-branch-${selected.Name}` ? "…" : "Rebase onto"}
                </button>
              )}
              {selected.IsLocal && (
                <button
                  onClick={() => setRenaming({ old: selected.Name, next: selected.Name })}
                  disabled={!!busy}
                  className="rounded-lg px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
                >
                  Rename
                </button>
              )}
              {selected.IsLocal && !selected.IsCurrent && (
                <button
                  onClick={() => setConfirmDel(selected.Name)}
                  disabled={!!busy}
                  className="rounded-lg px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10 disabled:opacity-40"
                >
                  {busy === `delbranch-${selected.Name}` ? "…" : "Delete"}
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      {confirmDel && (
        <ConfirmDialog
          title="Delete branch?"
          body={
            <>
              Delete branch <span className="text-foreground">{confirmDel}</span>? Safe delete
              refuses if it has unmerged commits.
            </>
          }
          confirmLabel="Force delete"
          busy={!!busy}
          onCancel={() => setConfirmDel(null)}
          onConfirm={() => doDelete(confirmDel, true)}
          extra={{ label: "Delete", onClick: () => doDelete(confirmDel, false) }}
        />
      )}
      {renaming && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-96 rounded-lg border border-border bg-[var(--color-background)] p-4 shadow-xl">
            <div className="mb-2 text-sm font-semibold">Rename branch</div>
            <input
              autoFocus
              value={renaming.next}
              onChange={(e) => setRenaming({ old: renaming.old, next: e.target.value })}
              className={`${cls.input} mb-4 w-full`}
            />
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setRenaming(null)}
                disabled={!!busy}
                className="rounded-lg border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40"
              >
                Cancel
              </button>
              <button
                onClick={doRename}
                disabled={!!busy || !renaming.next.trim() || renaming.next.trim() === renaming.old}
                className="rounded-lg bg-[var(--color-accent)] px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-40"
              >
                {busy === "rename-branch" ? "…" : "Rename"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}