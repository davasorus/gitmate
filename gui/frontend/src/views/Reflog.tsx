import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import type { ReflogEntry } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

// color the action word by kind
function actionColor(a: string): string {
  const k = a.split(" ")[0];
  if (k === "commit") return "text-[var(--color-added)]";
  if (k === "reset") return "text-[var(--color-removed)]";
  if (k === "rebase" || k === "merge") return "text-[var(--color-conflict)]";
  if (k === "checkout" || k === "switch") return "text-[var(--color-ahead)]";
  return "text-muted-foreground";
}

export function Reflog() {
  const { busy, setBusy, flash, service } = useGit();
  const [entries, setEntries] = useState<ReflogEntry[]>([]);

  const load = async () => {
    setBusy("reflog-load");
    try { setEntries((await service.Reflog(50)) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => { load(); }, []);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Reflog</h2>
        <button onClick={load} disabled={!!busy} className={cls.btnSm}>{busy === "reflog-load" ? "…" : "Refresh"}</button>
      </div>
      <div className="text-xs text-muted-foreground">
        Everywhere HEAD has been — the safety net. If a reset or rebase loses a commit, its hash is still here.
      </div>
      <div className="rounded-lg border border-border">
        {(entries ?? []).length ? (entries ?? []).map((e, i) => (
          <div key={i} className="flex items-baseline gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="shrink-0 text-[var(--color-modified)]">{e.Short}</span>
            <span className="shrink-0 text-xs text-muted-foreground">{e.Selector}</span>
            <span className={`shrink-0 text-xs ${actionColor(e.Action)}`}>{e.Action}</span>
            <span className="truncate text-xs">{e.Message}</span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no reflog entries</div>}
      </div>
    </div>
  );
}