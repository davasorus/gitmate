import { useEffect, useRef, useState } from "react";
import { useGit, cls } from "../context";
import type { WorkflowRun, Job, DispatchableWorkflow } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

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
  const { busy, setBusy, flash, service, run } = useGit();
  const [runs, setRuns] = useState<WorkflowRun[]>([]);
  const [openRun, setOpenRun] = useState<number | null>(null);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [filter, setFilter] = useState<StatusFilter>("all");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
  const [dispatchable, setDispatchable] = useState<DispatchableWorkflow[]>([]);
  // per-workflow run form (only shown for dispatchable workflows that declare inputs)
  const [runForm, setRunForm] = useState<number | null>(null); // WorkflowID
  const [runRef, setRunRef] = useState("live");
  const [runInputs, setRunInputs] = useState<Record<string, string>>({});
  const pollRef = useRef<number | null>(null);

  const loadRuns = async () => {
    setBusy("runs-load");
    try { setRuns((await service.ListRuns(50)) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const r = (await service.ListRuns(50)) ?? []; if (!cancelled) setRuns(r);
        const wfs = (await service.ListDispatchableWorkflows()) ?? []; if (!cancelled) setDispatchable(wfs);
      } catch (e) { if (!cancelled) flash("err", String(e)); }
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

  const doCancel = (r: WorkflowRun) => run(`cancel-${r.ID}`, async () => {
    await service.CancelRun(r.ID); await loadRuns(); return `cancelled run #${r.Number}`;
  }, "cancelled");
  const doRerun = (r: WorkflowRun, failedOnly: boolean) => run(`rerun-${r.ID}`, async () => {
    if (failedOnly) { await service.RerunFailed(r.ID); } else { await service.RerunRun(r.ID); }
    await loadRuns(); return `re-running #${r.Number}`;
  }, "re-run started");

  // dispatch: for a dispatchable workflow. No inputs → run immediately (the Run
  // button IS the run). Has inputs → reveal an inline form in that group, then run.
  const dispatchableFor = (workflowID: number) => dispatchable.find((d) => d.ID === workflowID);
  const startRun = (wf: DispatchableWorkflow) => {
    if (!wf.Inputs || wf.Inputs.length === 0) { doRunWorkflow(wf, "live", {}); return; }
    if (runForm === wf.ID) { setRunForm(null); return; } // toggle the form closed
    setRunForm(wf.ID); setRunRef("live"); setRunInputs({});
  };
  const doRunWorkflow = (wf: DispatchableWorkflow, ref: string, inputs: Record<string, string>) => {
    const file = wf.Path.split("/").pop() || wf.Path;
    run(`dispatch-${wf.ID}`, async () => {
      await service.TriggerDispatch(file, ref.trim() || "live", inputs);
      setRunForm(null);
      await loadRuns();
      return `dispatched ${wf.Name}`;
    }, `dispatched ${wf.Name}`);
  };

  // SMART POLLING: fixed 5s timer, only the open+active run, stops on completion/unmount.
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

  const filtered = (runs ?? []).filter((r) => matchesFilter(r, filter));
  const groups = new Map<string, { id: number; name: string; runs: WorkflowRun[] }>();
  for (const r of filtered) {
    const key = String(r.WorkflowID || r.WorkflowName || "other");
    if (!groups.has(key)) groups.set(key, { id: r.WorkflowID, name: r.WorkflowName || r.Name || "workflow", runs: [] });
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
          const wf = dispatchableFor(g.id);
          return (
            <div key={key} className="rounded-lg border border-border">
              <div className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm font-semibold">
                <button onClick={() => setCollapsed((c) => ({ ...c, [key]: !c[key] }))} className="flex flex-1 items-center gap-2 text-left hover:opacity-80">
                  <span className="text-muted-foreground">{isCollapsed ? "▸" : "▾"}</span>
                  <span className="truncate">{g.name}</span>
                  {anyActive && <span className="text-[10px] text-[var(--color-behind)]">live</span>}
                  <span className="ml-auto text-xs font-normal text-muted-foreground">{g.runs.length}</span>
                </button>
                {wf && (
                  <button onClick={() => startRun(wf)} disabled={!!busy}
                          className={cls.btnSm} title="run this workflow (workflow_dispatch)">
                    {busy === `dispatch-${wf.ID}` ? "…" : "Run"}
                  </button>
                )}
              </div>

              {/* inline dispatch form — only when this workflow has inputs and Run was clicked */}
              {wf && runForm === wf.ID && (
                <div className="space-y-2 border-b border-border bg-background px-3 py-2">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">ref</span>
                    <input value={runRef} onChange={(e) => setRunRef(e.target.value)} placeholder="branch or tag" className={`${cls.input} flex-1 text-xs`} />
                  </div>
                  {(wf.Inputs ?? []).map((inp) => (
                    <div key={inp.Name} className="space-y-1">
                      <label className="text-xs">{inp.Name}{inp.Required && <span className="text-[var(--color-removed)]"> *</span>}{inp.Description ? <span className="text-muted-foreground"> — {inp.Description}</span> : null}</label>
                      {inp.Type === "choice" ? (
                        <select value={runInputs[inp.Name] ?? inp.Default} onChange={(e) => setRunInputs((m) => ({ ...m, [inp.Name]: e.target.value }))} className={`${cls.input} w-full text-xs`}>
                          {(inp.Options ?? []).map((o) => <option key={o} value={o}>{o}</option>)}
                        </select>
                      ) : inp.Type === "boolean" ? (
                        <select value={runInputs[inp.Name] ?? inp.Default ?? "false"} onChange={(e) => setRunInputs((m) => ({ ...m, [inp.Name]: e.target.value }))} className={`${cls.input} w-full text-xs`}>
                          <option value="true">true</option><option value="false">false</option>
                        </select>
                      ) : (
                        <input value={runInputs[inp.Name] ?? inp.Default} onChange={(e) => setRunInputs((m) => ({ ...m, [inp.Name]: e.target.value }))} className={`${cls.input} w-full text-xs`} />
                      )}
                    </div>
                  ))}
                  <div className="flex justify-end gap-2">
                    <button onClick={() => setRunForm(null)} disabled={!!busy} className="rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40">Cancel</button>
                    <button onClick={() => doRunWorkflow(wf, runRef, runInputs)} disabled={!!busy} className={cls.btnSm}>{busy === `dispatch-${wf.ID}` ? "…" : "Run"}</button>
                  </div>
                </div>
              )}

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
                      <div className="flex items-center gap-1 pt-1">
                        {isActive(r)
                          ? <button onClick={() => doCancel(r)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)]`}>{busy === `cancel-${r.ID}` ? "…" : "Cancel"}</button>
                          : <>
                              <button onClick={() => doRerun(r, false)} disabled={!!busy} className={cls.btnSm}>{busy === `rerun-${r.ID}` ? "…" : "Re-run"}</button>
                              {r.Conclusion === "failure" && <button onClick={() => doRerun(r, true)} disabled={!!busy} className={cls.btnSm}>Re-run failed</button>}
                            </>}
                        <a href={r.URL} target="_blank" rel="noreferrer" className="ml-auto text-[10px] text-[var(--color-ahead)] underline">View on GitHub</a>
                      </div>
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