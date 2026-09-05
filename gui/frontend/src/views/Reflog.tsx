import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { ReflogEntry } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

function actionColor(a: string): string {
  const k = a.split(" ")[0];
  if (k === "commit") return "text-[var(--color-added)]";
  if (k === "reset") return "text-[var(--color-removed)]";
  if (k === "rebase" || k === "merge") return "text-[var(--color-conflict)]";
  if (k === "checkout" || k === "switch") return "text-[var(--color-ahead)]";
  return "text-muted-foreground";
}

export function Reflog() {
  const { busy, setBusy, flash, run, service, reload } = useGit();
  const [entries, setEntries] = useState<ReflogEntry[]>([]);
  const [confirmHard, setConfirmHard] = useState<ReflogEntry | null>(null);

  const load = async () => {
    setBusy("reflog-load");
    try { setEntries((await service.Reflog(50)) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => { load(); }, []);

  const doReset = (rev: string, mode: "soft" | "mixed" | "hard") =>
    run(`reset-${rev}`, async () => {
      await service.Reset(rev, mode);
      setConfirmHard(null);
      await load();
      return `reset (${mode}) to ${rev}`;
    }, `reset to ${rev}`);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Reflog</h2>
        <button onClick={load} disabled={!!busy} className={cls.btnSm}>{busy === "reflog-load" ? "…" : "Refresh"}</button>
      </div>
      <div className="text-xs text-muted-foreground">
        Everywhere HEAD has been. "Reset to here" moves HEAD back to that point — soft keeps your changes staged, mixed keeps them unstaged, hard discards them (recoverable from this very log).
      </div>
      <div className="rounded-lg border border-border">
        {(entries ?? []).length ? (entries ?? []).map((e, i) => (
          <div key={i} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="shrink-0 text-[var(--color-modified)]">{e.Short}</span>
            <span className="shrink-0 text-xs text-muted-foreground">{e.Selector}</span>
            <span className={`shrink-0 text-xs ${actionColor(e.Action)}`}>{e.Action}</span>
            <span className="truncate text-xs">{e.Message}</span>
            <span className="ml-auto flex shrink-0 gap-1">
              <button onClick={() => doReset(e.Short, "soft")} disabled={!!busy} className={cls.btnSm} title="move HEAD here, keep changes staged">
                {busy === `reset-${e.Short}` ? "…" : "Soft"}
              </button>
              <button onClick={() => doReset(e.Short, "mixed")} disabled={!!busy} className={cls.btnSm} title="move HEAD here, keep changes unstaged">
                Mixed
              </button>
              <button onClick={() => setConfirmHard(e)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`} title="move HEAD here, DISCARD changes">
                Hard
              </button>
            </span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no reflog entries</div>}
      </div>

      {confirmHard && (
        <ConfirmDialog title="Hard reset?"
          body={<>Move HEAD to <span className="text-foreground">{confirmHard.Short}</span> and <b>discard</b> all uncommitted changes and any commits after it. Recoverable from the reflog until it expires — but not from your working tree.</>}
          confirmLabel="Hard reset" busy={!!busy}
          onCancel={() => setConfirmHard(null)}
          onConfirm={() => doReset(confirmHard.Short, "hard")} />
      )}
    </div>
  );
}