import { useEffect, useRef, useState } from "react";
import { useGit, cls } from "../context";
import type { WorkflowRun, Job } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

type StatusFilter = "all" | "success" | "failure" | "in_progress";

function statusColor(status: string, conclusion: string): string {
  const s = conclusion || status;
  if (s === "success") return "text-[var(--color-added)]";
  if (s === "failure" || s === "cancelled" || s === "timed_out") return "text-[var(--color-removed)]";
  if (s === "in_progress" || s === "queued" || s === "pending") return "text-[var(--color-behind)]";
  return "text-muted-foreground";
}
const statusLabel = (status: string, conclusion: string) => conclusion || status || "—";
const isActive = (r: WorkflowRun) => r.Status === "in_progress" || r.Status === "queued";

function matchesFilter(r: WorkflowRun, f: StatusFilter): boolean {
  if (f === "all") return true;
  if (f === "in_progress") return isActive(r);
  return r.Conclusion === f;
}

export function Actions() {
  const { busy, setBusy, flash, service } = useGit();
  const [runs, setRuns] = useState<WorkflowRun[]>([]);
  const [openRun, setOpenRun] = useState<number | null>(null);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [filter, setFilter] = useState<StatusFilter>("all");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
  const pollRef = useRef<number | null>(null);

  const loadRuns = async () => {
    setBusy("runs-load");
    try { setRuns((await service.ListRuns(50)) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try { const r = (await service.ListRuns(50)) ?? []; if (!cancelled) setRuns(r); }
      catch (e) { if (!cancelled) flash("err", String(e)); }
    })();
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const loadJobs = async (runID: number) => {
    try { setJobs((await service.RunJobs(runID)) ?? []); }
    catch (e) { flash("err", String(e)); }
  };
  const toggleRun = (r: WorkflowRun) => {
    if (openRun === r.ID) { setOpenRun(null); setJobs([]); return; }
    setOpenRun(r.ID); loadJobs(r.ID);
  };

  // SMART POLLING (unchanged from C-1): fixed 5s timer, only the open+active run,
  // stops on completion/unmount, timer-driven not render-driven.
  useEffect(() => {
    if (openRun === null) return;
    const o = runs.find((r) => r.ID === openRun);
    if (!o || !isActive(o)) return;
    pollRef.current = window.setInterval(async () => {
      try {
        const fresh = await service.GetRun(openRun);
        setRuns((prev) => prev.map((r) => (r.ID === openRun ? fresh : r)));
        setJobs((await service.RunJobs(openRun)) ?? []);
        if (!(fresh.Status === "in_progress" || fresh.Status === "queued")) {
          if (pollRef.current) { clearInterval(pollRef.current); pollRef.current = null; }
        }
      } catch { /* transient */ }
    }, 5000);
    return () => { if (pollRef.current) { clearInterval(pollRef.current); pollRef.current = null; } };
  }, [openRun, runs, service]);

  // group filtered runs by workflow (WorkflowID key, WorkflowName label)
  const filtered = (runs ?? []).filter((r) => matchesFilter(r, filter));
  const groups = new Map<string, { name: string; runs: WorkflowRun[] }>();
  for (const r of filtered) {
    const key = String(r.WorkflowID || r.WorkflowName || "other");
    if (!groups.has(key)) groups.set(key, { name: r.WorkflowName || r.Name || "workflow", runs: [] });
    groups.get(key)!.runs.push(r);
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Actions</h2>
        <div className="flex items-center gap-1 text-xs">
          {(["all", "success", "failure", "in_progress"] as StatusFilter[]).map((f) => (
            <button key={f} onClick={() => setFilter(f)}
                    className={`rounded-md border border-border px-2 py-0.5 ${filter === f ? "bg-muted font-semibold" : "hover:bg-muted/60"}`}>
              {f === "in_progress" ? "running" : f}
            </button>
          ))}
          <button onClick={loadRuns} disabled={!!busy} className={cls.btnSm}>{busy === "runs-load" ? "…" : "Refresh"}</button>
        </div>
      </div>

      {groups.size === 0 ? (
        <div className="rounded-lg border border-border p-3 text-sm italic text-muted-foreground">no runs match</div>
      ) : (
        Array.from(groups.entries()).map(([key, g]) => {
          const isCollapsed = collapsed[key];
          const anyActive = g.runs.some(isActive);
          return (
            <div key={key} className="rounded-lg border border-border">
              <button onClick={() => setCollapsed((c) => ({ ...c, [key]: !c[key] }))}
                      className="flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-sm font-semibold hover:bg-muted/60">
                <span className="text-muted-foreground">{isCollapsed ? "▸" : "▾"}</span>
                <span className="truncate">{g.name}</span>
                {anyActive && <span className="text-[10px] text-[var(--color-behind)]">live</span>}
                <span className="ml-auto text-xs font-normal text-muted-foreground">{g.runs.length}</span>
              </button>
              {!isCollapsed && g.runs.map((r) => (
                <div key={r.ID} className="border-b border-border last:border-0">
                  <button onClick={() => toggleRun(r)} className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted/60">
                    <span className={`shrink-0 ${statusColor(r.Status, r.Conclusion)}`}>●</span>
                    <span className="shrink-0 text-muted-foreground">#{r.Number}</span>
                    <span className="shrink-0 text-[var(--color-ahead)]">{r.Branch}</span>
                    <span className="shrink-0 text-muted-foreground">{r.Event}</span>
                    {r.Duration && <span className="shrink-0 text-muted-foreground/70">{r.Duration}</span>}
                    <span className="shrink-0 text-muted-foreground/60">{r.CreatedAt}</span>
                    <span className={`ml-auto shrink-0 ${statusColor(r.Status, r.Conclusion)}`}>{statusLabel(r.Status, r.Conclusion)}</span>
                  </button>
                  {openRun === r.ID && (
                    <div className="space-y-1 bg-background px-3 pb-2 pt-1">
                      {(jobs ?? []).length ? (jobs ?? []).map((j) => (
                        <div key={j.ID}>
                          <div className="flex items-center gap-2 text-xs">
                            <span className={statusColor(j.Status, j.Conclusion)}>●</span>
                            <span className="font-medium">{j.Name}</span>
                            <span className={`text-[10px] ${statusColor(j.Status, j.Conclusion)}`}>{statusLabel(j.Status, j.Conclusion)}</span>
                          </div>
                          <div className="ml-4">
                            {(j.Steps ?? []).map((st, si) => (
                              <div key={si} className="flex items-center gap-2 text-[11px]">
                                <span className={statusColor(st.Status, st.Conclusion)}>○</span>
                                <span className="truncate">{st.Name}</span>
                                <span className={`text-[10px] ${statusColor(st.Status, st.Conclusion)}`}>{statusLabel(st.Status, st.Conclusion)}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )) : <div className="text-xs italic text-muted-foreground">no jobs (or still starting)</div>}
                      <a href={r.URL} target="_blank" rel="noreferrer" className="text-[10px] text-[var(--color-ahead)] underline">View on GitHub</a>
                    </div>
                  )}
                </div>
              ))}
            </div>
          );
        })
      )}
    </div>
  );
}