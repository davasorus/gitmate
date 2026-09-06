import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { DiffView } from "../components/DiffView";
import type {
  Commit,
  CommitDetail,
} from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function History() {
  const { commits, branches, status, busy, setBusy, flash, service, run } = useGit();
  const [openCommit, setOpenCommit] = useState<string | null>(null);
  const [detail, setDetail] = useState<CommitDetail | null>(null);

  const [viewBranch, setViewBranch] = useState(""); // "" = current branch
  const [refCommits, setRefCommits] = useState<Commit[] | null>(null);

  const current = status?.Branch ?? "";
  const shown = refCommits ?? commits ?? [];
  const viewingOther = viewBranch !== "" && viewBranch !== current;

  const loadBranch = async (ref: string) => {
    setViewBranch(ref);
    if (ref === "" || ref === current) {
      setRefCommits(null);
      return;
    }
    setBusy("history-branch");
    try {
      setRefCommits((await service.LogRef(ref, 30)) ?? []);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  // reset to current branch when it changes (e.g. after a switch)
  useEffect(() => {
    setViewBranch("");
    setRefCommits(null);
  }, [current]);

  const showCommit = async (hash: string) => {
    if (openCommit === hash) {
      setOpenCommit(null);
      setDetail(null);
      return;
    }
    setBusy(`show-${hash}`);
    try {
      setDetail(await service.Show(hash));
      setOpenCommit(hash);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const doCherryPick = (hash: string) =>
    run(
      `cherry-${hash}`,
      async () => {
        await service.CherryPick(hash);
        const c = (await service.ConflictedFiles()) ?? [];
        return c.length
          ? `cherry-pick hit ${c.length} conflict(s) — see banner`
          : `cherry-picked ${hash.slice(0, 7)} onto ${current}`;
      },
      `cherry-picked ${hash.slice(0, 7)}`,
    );
  const doRevert = (hash: string) =>
    run(
      `revert-${hash}`,
      async () => {
        await service.Revert(hash);
        const c = (await service.ConflictedFiles()) ?? [];
        return c.length
          ? `revert hit ${c.length} conflict(s) — see banner`
          : `reverted ${hash.slice(0, 7)}`;
      },
      `reverted ${hash.slice(0, 7)}`,
    );

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
          History
        </h2>
        <div className="flex items-center gap-2 text-xs">
          <span className="text-muted-foreground">branch:</span>
          <select
            value={viewBranch || current}
            onChange={(e) => loadBranch(e.target.value)}
            className={cls.input}
          >
            {(branches ?? []).map((b) => (
              <option key={b.Name} value={b.Name}>
                {b.Name}
                {b.IsCurrent ? " (current)" : ""}
              </option>
            ))}
          </select>
        </div>
      </div>

      {viewingOther && (
        <div className="rounded-md border border-[var(--color-ahead)]/40 bg-[var(--color-ahead)]/5 px-3 py-1.5 text-xs text-muted-foreground">
          Viewing <span className="text-[var(--color-ahead)]">{viewBranch}</span>. Cherry-pick /
          revert here apply onto your current branch{" "}
          <span className="text-[var(--color-added)]">{current}</span>.
        </div>
      )}

      <div className="rounded-lg border border-border">
        {busy === "history-branch" ? (
          <div className="p-3 text-sm text-muted-foreground">…</div>
        ) : shown.length === 0 ? (
          <div className="p-3 text-sm italic text-muted-foreground">no commits</div>
        ) : (
          shown.map((c) => (
            <div key={c.Hash}>
              <div className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0 hover:bg-muted/60">
                <button
                  onClick={() => showCommit(c.Hash)}
                  className="flex flex-1 items-baseline gap-2 text-left"
                >
                  <span className="shrink-0 text-[var(--color-modified)]">{c.Short}</span>
                  <span className="truncate">{c.Subject}</span>
                  <span className="ml-auto shrink-0 text-xs text-muted-foreground">{c.Author}</span>
                </button>
                <span className="flex shrink-0 gap-1">
                  <button
                    onClick={() => doCherryPick(c.Hash)}
                    disabled={!!busy}
                    className="rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40"
                    title={`apply this commit onto ${current}`}
                  >
                    {busy === `cherry-${c.Hash}` ? "…" : "Cherry-pick"}
                  </button>
                  <button
                    onClick={() => doRevert(c.Hash)}
                    disabled={!!busy}
                    className="rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40"
                    title="create a commit undoing this one"
                  >
                    {busy === `revert-${c.Hash}` ? "…" : "Revert"}
                  </button>
                </span>
              </div>
              {openCommit === c.Hash && detail && (
                <div className="border-b border-border bg-background px-2 py-2">
                  <div className="mb-2 px-1 text-xs text-muted-foreground">
                    <div>
                      <span className="text-foreground">{detail.Subject}</span>
                    </div>
                    <div>
                      {detail.Author} &lt;{detail.Email}&gt; · {detail.Date}
                    </div>
                    {detail.Body ? (
                      <div className="mt-1 whitespace-pre-wrap">{detail.Body}</div>
                    ) : null}
                  </div>
                  <DiffView files={detail.Files ?? []} />
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
