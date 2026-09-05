import { useState } from "react";
import { useGit } from "../context";
import { DiffView } from "../components/DiffView";
import type { CommitDetail } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function History() {
  const { commits, busy, setBusy, flash, service, run } = useGit();
  const [openCommit, setOpenCommit] = useState<string | null>(null);
  const [detail, setDetail] = useState<CommitDetail | null>(null);

  const showCommit = async (hash: string) => {
    if (openCommit === hash) { setOpenCommit(null); setDetail(null); return; }
    setBusy(`show-${hash}`);
    try { setDetail(await service.Show(hash)); setOpenCommit(hash); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };

  const doCherryPick = (hash: string) =>
    run(`cherry-${hash}`, async () => {
      await service.CherryPick(hash);
      const c = (await service.ConflictedFiles()) ?? [];
      return c.length ? `cherry-pick hit ${c.length} conflict(s) — see banner` : `cherry-picked ${hash.slice(0,7)}`;
    }, `cherry-picked ${hash.slice(0,7)}`);
  const doRevert = (hash: string) =>
    run(`revert-${hash}`, async () => {
      await service.Revert(hash);
      const c = (await service.ConflictedFiles()) ?? [];
      return c.length ? `revert hit ${c.length} conflict(s) — see banner` : `reverted ${hash.slice(0,7)}`;
    }, `reverted ${hash.slice(0,7)}`);

  return (
    <div className="space-y-2">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">History</h2>
      <div className="rounded-lg border border-border">
        {(commits ?? []).map((c) => (
          <div key={c.Hash}>
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0 hover:bg-muted/60">
              <button onClick={() => showCommit(c.Hash)} className="flex flex-1 items-baseline gap-2 text-left">
                <span className="shrink-0 text-[var(--color-modified)]">{c.Short}</span>
                <span className="truncate">{c.Subject}</span>
                <span className="ml-auto shrink-0 text-xs text-muted-foreground">{c.Author}</span>
              </button>
              <span className="flex shrink-0 gap-1">
                <button onClick={() => doCherryPick(c.Hash)} disabled={!!busy} className="rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40" title="apply this commit onto the current branch">
                  {busy === `cherry-${c.Hash}` ? "…" : "Cherry-pick"}
                </button>
                <button onClick={() => doRevert(c.Hash)} disabled={!!busy} className="rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40" title="create a commit undoing this one">
                  {busy === `revert-${c.Hash}` ? "…" : "Revert"}
                </button>
              </span>
            </div>
            {openCommit === c.Hash && detail && (
              <div className="border-b border-border bg-background px-2 py-2">
                <div className="mb-2 px-1 text-xs text-muted-foreground">
                  <div><span className="text-foreground">{detail.Subject}</span></div>
                  <div>{detail.Author} &lt;{detail.Email}&gt; · {detail.Date}</div>
                  {detail.Body ? <div className="mt-1 whitespace-pre-wrap">{detail.Body}</div> : null}
                </div>
                <DiffView files={detail.Files} />
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}