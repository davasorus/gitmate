import { useState } from "react";
import { useGit, cls } from "../context";
import type { ConflictFile } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function Conflicts() {
  const { conflicts, mergeInProgress, busy, run, service, flash, setBusy } = useGit();
  const [open, setOpen] = useState<string | null>(null);
  const [detail, setDetail] = useState<ConflictFile | null>(null);

  const view = async (path: string) => {
    if (open === path) { setOpen(null); setDetail(null); return; }
    setBusy(`conflict-${path}`);
    try { setDetail(await service.ReadConflict(path)); setOpen(path); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };

  const takeOurs = (p: string) => run(`ours-${p}`, async () => { await service.ResolveOurs(p); setOpen(null); return `took ours for ${p}`; }, "resolved");
  const takeTheirs = (p: string) => run(`theirs-${p}`, async () => { await service.ResolveTheirs(p); setOpen(null); return `took theirs for ${p}`; }, "resolved");
  const markResolved = (p: string) => run(`mark-${p}`, async () => { await service.MarkResolved(p); setOpen(null); return `marked ${p} resolved`; }, "resolved");
  const abort = () => run("merge-abort", () => service.MergeAbort(), "merge aborted");

  if (!mergeInProgress) {
    return (
      <div className="space-y-2">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Conflicts</h2>
        <div className="rounded-lg border border-border p-3 text-sm italic text-muted-foreground">
          No merge in progress. Conflicts appear here when a merge, rebase, or pull hits them.
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Conflicts</h2>
        <button onClick={abort} disabled={!!busy} className={cls.btnSm}>{busy === "merge-abort" ? "…" : "Abort merge"}</button>
      </div>

      {conflicts.length === 0 ? (
        <div className="rounded-md border border-[var(--color-added)] bg-[var(--color-added)]/10 px-3 py-2 text-sm text-[var(--color-added)]">
          All conflicts resolved — commit in Changes to finish the merge.
        </div>
      ) : (
        <div className="rounded-lg border border-border">
          {conflicts.map((path) => (
            <div key={path}>
              <div className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
                <span className="shrink-0 text-[var(--color-conflict)]">conflict</span>
                <button onClick={() => view(path)} className="flex-1 truncate text-left hover:underline">{path}</button>
                <span className="flex shrink-0 gap-1">
                  <button onClick={() => takeOurs(path)} disabled={!!busy} className={cls.btnSm}>{busy === `ours-${path}` ? "…" : "Take ours"}</button>
                  <button onClick={() => takeTheirs(path)} disabled={!!busy} className={cls.btnSm}>{busy === `theirs-${path}` ? "…" : "Take theirs"}</button>
                  <button onClick={() => markResolved(path)} disabled={!!busy} className={cls.btnSm}>{busy === `mark-${path}` ? "…" : "Mark resolved"}</button>
                </span>
              </div>
              {open === path && detail && (
                <div className="space-y-2 border-b border-border bg-background px-3 py-2">
                  {(detail.Hunks ?? []).length === 0 && <div className="text-xs italic text-muted-foreground">no conflict regions found (file may already be resolved or hand-edited)</div>}
                  {(detail.Hunks ?? []).map((h, i) => (
                    <div key={i} className="grid grid-cols-2 gap-2 text-xs">
                      <div className="rounded border border-[var(--color-added)]/40 bg-[var(--color-added)]/5 p-2">
                        <div className="mb-1 font-semibold text-[var(--color-added)]">ours — current branch (HEAD)</div>
                        <pre className="whitespace-pre-wrap">{(h.Ours ?? []).join("\n")}</pre>
                      </div>
                      <div className="rounded border border-[var(--color-ahead)]/40 bg-[var(--color-ahead)]/5 p-2">
                        <div className="mb-1 font-semibold text-[var(--color-ahead)]">theirs — merged-in branch</div>
                        <pre className="whitespace-pre-wrap">{(h.Theirs ?? []).join("\n")}</pre>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}