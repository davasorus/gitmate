import { useEffect, useState } from "react";
import { useGit } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { ReflogEntry } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

function actionColor(a: string): string {
  const k = a.split(" ")[0];
  if (k === "commit") return "text-[var(--color-added)]";
  if (k === "reset") return "text-[var(--color-removed)]";
  if (k === "rebase" || k === "merge") return "text-[var(--color-conflict)]";
  if (k === "checkout" || k === "switch") return "text-[var(--color-accent)]";
  return "text-[var(--color-faint)]";
}

export function Reflog() {
  const { busy, setBusy, flash, run, service } = useGit();
  const [entries, setEntries] = useState<ReflogEntry[]>([]);
  const [confirmHard, setConfirmHard] = useState<ReflogEntry | null>(null);

  const load = async () => {
    setBusy("reflog-load");
    try {
      setEntries((await service.Reflog(50)) ?? []);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };
  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const doReset = (rev: string, mode: "soft" | "mixed" | "hard") =>
    run(
      `reset-${rev}`,
      async () => {
        await service.Reset(rev, mode);
        setConfirmHard(null);
        await load();
        return `reset (${mode}) to ${rev}`;
      },
      `reset to ${rev}`,
    );

  return (
    <div className="mx-auto max-w-3xl space-y-3">
      <div className="flex items-center gap-2">
        <span className="text-[15px] font-semibold tracking-[-0.01em]">Reflog</span>
        <button
          onClick={load}
          disabled={!!busy}
          className="ml-auto rounded-lg border border-border px-2.5 py-1 text-[12px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
        >
          {busy === "reflog-load" ? "…" : "Refresh"}
        </button>
      </div>
      <div className="text-[12px] text-[var(--color-muted-foreground)]">
        Everywhere HEAD has been. Reset moves HEAD back — soft keeps changes staged, mixed keeps
        them unstaged, hard discards them (recoverable from this log).
      </div>

      <div className="space-y-1">
        {(entries ?? []).length ? (
          entries.map((e, i) => (
            <div
              key={i}
              className="group flex items-center gap-2.5 rounded-lg border border-border px-3 py-2 hover:bg-[var(--color-card)]"
            >
              <span className="shrink-0 font-mono text-[11px] text-[var(--color-modified)]">
                {e.Short}
              </span>
              <span className="shrink-0 font-mono text-[10.5px] text-[var(--color-faint)]">
                {e.Selector}
              </span>
              <span className={`shrink-0 text-[11px] ${actionColor(e.Action)}`}>{e.Action}</span>
              <span className="min-w-0 flex-1 truncate text-[12.5px]">{e.Message}</span>
              <span className="flex shrink-0 gap-1 opacity-0 group-hover:opacity-100">
                <button
                  onClick={() => doReset(e.Short, "soft")}
                  disabled={!!busy}
                  title="move HEAD here, keep changes staged"
                  className="rounded-md px-2 py-0.5 text-[11px] text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)]"
                >
                  {busy === `reset-${e.Short}` ? "…" : "Soft"}
                </button>
                <button
                  onClick={() => doReset(e.Short, "mixed")}
                  disabled={!!busy}
                  title="move HEAD here, keep changes unstaged"
                  className="rounded-md px-2 py-0.5 text-[11px] text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)]"
                >
                  Mixed
                </button>
                <button
                  onClick={() => setConfirmHard(e)}
                  disabled={!!busy}
                  title="move HEAD here, DISCARD changes"
                  className="rounded-md px-2 py-0.5 text-[11px] text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10"
                >
                  Hard
                </button>
              </span>
            </div>
          ))
        ) : (
          <div className="rounded-lg border border-border px-3 py-6 text-center text-[12.5px] italic text-[var(--color-faint)]">
            no reflog entries
          </div>
        )}
      </div>

      {confirmHard && (
        <ConfirmDialog
          title="Hard reset?"
          body={
            <>
              Move HEAD to <span className="text-foreground">{confirmHard.Short}</span> and{" "}
              <b>discard</b> all uncommitted changes and any commits after it. Recoverable from the
              reflog until it expires — but not from your working tree.
            </>
          }
          confirmLabel="Hard reset"
          busy={!!busy}
          onCancel={() => setConfirmHard(null)}
          onConfirm={() => doReset(confirmHard.Short, "hard")}
        />
      )}
    </div>
  );
}
