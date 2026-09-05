import { useState } from "react";
import { useGit } from "../context";
import { DiffView } from "../components/DiffView";
import type { CommitDetail } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function History() {
  const { commits, busy, setBusy, flash, service } = useGit();
  const [openCommit, setOpenCommit] = useState<string | null>(null);
  const [detail, setDetail] = useState<CommitDetail | null>(null);

  const showCommit = async (hash: string) => {
    if (openCommit === hash) { setOpenCommit(null); setDetail(null); return; }
    setBusy(`show-${hash}`);
    try { setDetail(await service.Show(hash)); setOpenCommit(hash); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };

  return (
    <div className="space-y-2">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">History</h2>
      <div className="rounded-lg border border-border">
        {(commits ?? []).map((c) => (
          <div key={c.Hash}>
            <button onClick={() => showCommit(c.Hash)}
                    className="flex w-full items-baseline gap-2 border-b border-border px-3 py-1.5 text-left text-sm last:border-0 hover:bg-muted/60">
              <span className="shrink-0 text-[var(--color-modified)]">{c.Short}</span>
              <span className="truncate">{c.Subject}</span>
              <span className="ml-auto shrink-0 text-xs text-muted-foreground">{c.Author}</span>
            </button>
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
