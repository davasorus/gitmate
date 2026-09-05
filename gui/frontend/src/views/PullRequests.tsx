import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { CheckBadge } from "../components/CheckBadge";
import type { CheckRun } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

export function PullRequests() {
  const { prs, branches, busy, run, service, flash, setBusy } = useGit();
  const [prTitle, setPrTitle] = useState("");
  const [prHead, setPrHead] = useState("");
  const [prBase, setPrBase] = useState("live");
  const [prBody, setPrBody] = useState("");
  const [prResult, setPrResult] = useState("");
  const [checks, setChecks] = useState<Record<number, CheckRun[]>>({});

  // prefill head with current branch + body with template, once
  useEffect(() => {
    (async () => {
      try {
        const cur = await service.CurrentBranch();
        setPrHead((h) => h || cur);
        const tmpl = await service.PRTemplate();
        setPrBody((b) => b || tmpl);
      } catch { /* non-fatal */ }
    })();
  }, []);

  const doPR = () => run("pr", async () => {
    let title = prTitle.trim();
    if (!title) title = (await service.DefaultPRTitle(prHead)) || prHead;
    const url = await service.CreatePR(title, prBody, prHead, prBase);
    setPrResult(url); setPrTitle("");
    return url;
  }, "PR opened");
  const doMerge = (n: number) => run(`merge-${n}`, async () => { const sha = await service.MergePR(n, "merge"); return `merged #${n} (${sha.slice(0, 7)})`; }, `merged #${n}`);
 const loadChecks = async (n: number) => {
    setBusy(`checks-${n}`);
    try {
      const runs = (await service.PRChecks(n)) ?? [];
      setChecks((p) => ({ ...p, [n]: runs }));
    } catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Pull requests</h2>

      <div className="space-y-2 rounded-lg border border-border p-3">
        <div className="text-xs text-muted-foreground">
          Open a PR from <span className="text-[var(--color-ahead)]">{prHead || "…"}</span> into <span className="text-[var(--color-ahead)]">{prBase}</span>
        </div>
        <div className="flex flex-wrap gap-2">
          <input value={prHead} onChange={(e) => setPrHead(e.target.value)} placeholder="head branch" className={`${cls.input} flex-1`} />
          <select value={prBase} onChange={(e) => setPrBase(e.target.value)} className={cls.input}>
            {(branches ?? []).map((b) => <option key={b.Name} value={b.Name}>{b.Name}</option>)}
            {!(branches ?? []).some((b) => b.Name === prBase) && <option value={prBase}>{prBase}</option>}
          </select>
        </div>
        <input value={prTitle} onChange={(e) => setPrTitle(e.target.value)} placeholder="title (optional — defaults to last commit subject)" className={`${cls.input} w-full`} />
        <textarea value={prBody} onChange={(e) => setPrBody(e.target.value)} placeholder="description (loaded from PR template if present)" className={`${cls.input} h-28 w-full resize-y`} />
        <div className="flex items-center gap-3">
          <button onClick={doPR} disabled={!!busy || !prHead.trim()} className={cls.btn}>{busy === "pr" ? "…" : "Open PR"}</button>
          {prResult && <a href={prResult} target="_blank" rel="noreferrer" className="text-xs text-[var(--color-ahead)] underline">{prResult}</a>}
        </div>
      </div>

      <div className="rounded-lg border border-border">
        {(prs ?? []).length ? (prs ?? []).map((p) => (
          <div key={p.Number} className="border-b border-border p-3 last:border-0">
            <div className="flex items-baseline gap-2 text-sm">
              <span className="font-semibold text-[var(--color-ahead)]">#{p.Number}</span>
              <span className="truncate">{p.Title}</span>
              <span className="text-xs text-muted-foreground">@{p.Author}</span>
              <span className="ml-auto flex gap-1">
                <button onClick={() => loadChecks(p.Number)} disabled={!!busy} className={cls.btnSm}>{busy === `checks-${p.Number}` ? "…" : "Checks"}</button>
                <button onClick={() => doMerge(p.Number)} disabled={!!busy} className={cls.btnSm}>{busy === `merge-${p.Number}` ? "…" : "Merge"}</button>
              </span>
            </div>
            {checks[p.Number] && (
              <div className="mt-1 space-y-0.5 pl-6">
                {checks[p.Number].length ? checks[p.Number].map((r, i) => <CheckBadge key={i} run={r} />) : <div className="text-xs italic text-muted-foreground">no checks reported</div>}
              </div>
            )}
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no open PRs</div>}
      </div>
    </div>
  );
}
