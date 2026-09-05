import { createContext, useContext, useEffect, useState, useCallback, ReactNode } from "react";
import { GitService } from "../bindings/github.com/davasorus/gitmate/gui";
import type { Status, Commit, Branch, FileDiff, Stash, CommitDetail } from "../bindings/github.com/davasorus/gitmate/internal/gitops";
import type { PR, CheckRun, Issue } from "../bindings/github.com/davasorus/gitmate/internal/ghapi";

export type View = "changes" | "history" | "branches" | "prs" | "issues" | "stashes";
type Toast = { kind: "ok" | "err"; msg: string } | null;

export interface GitmateState {
  // view + repo
  view: View; setView: (v: View) => void;
  dir: string; setDir: (d: string) => void;

  // data
  status: Status | null;
  branches: Branch[];
  commits: Commit[];
  prs: PR[];
  issues: Issue[];
  stashes: Stash[];
  checks: Record<number, CheckRun[]>;
  mergeInProgress: boolean;
  conflicts: string[];

  // ui
  toast: Toast;
  busy: string;
  setBusy: (b: string) => void;
  flash: (kind: "ok" | "err", msg: string) => void;
  reload: () => Promise<void>;
  run: (name: string, fn: () => Promise<string | void>, okMsg: string) => Promise<void>;

  // diff cache (Changes + History share via callers)
  service: typeof GitService;
}

const Ctx = createContext<GitmateState | null>(null);
export const useGit = () => {
  const c = useContext(Ctx);
  if (!c) throw new Error("useGit must be used within GitmateProvider");
  return c;
};

export function GitmateProvider({ children }: { children: ReactNode }) {
  const [view, setView] = useState<View>("changes");
  const [dir, setDir] = useState("");

  const [status, setStatus] = useState<Status | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [prs, setPRs] = useState<PR[]>([]);
  const [issues, setIssues] = useState<Issue[]>([]);
  const [stashes, setStashes] = useState<Stash[]>([]);
  const [checks] = useState<Record<number, CheckRun[]>>({});
  const [mergeInProgress, setMergeInProgress] = useState(false);
  const [conflicts, setConflicts] = useState<string[]>([]);

  const [toast, setToast] = useState<Toast>(null);
  const [busy, setBusy] = useState("");

  const flash = (kind: "ok" | "err", msg: string) => {
    setToast({ kind, msg });
    setTimeout(() => setToast(null), 4000);
  };

  const reload = useCallback(async () => {
    try {
      await GitService.SetRepoDir(dir.trim());
      const [s, b, c] = await Promise.all([
        GitService.Status(),
        GitService.Branches(),
        GitService.Log(15),
      ]);
      setStatus(s);
      setBranches(b ?? []);
      setCommits(c ?? []);
      try { setPRs((await GitService.PRs("open")) ?? []); } catch { setPRs([]); }
      try { setStashes((await GitService.StashList()) ?? []); } catch { setStashes([]); }
      try { setIssues((await GitService.Issues("open")) ?? []); } catch { setIssues([]); }
      try {
        setMergeInProgress(await GitService.MergeInProgress());
        setConflicts((await GitService.ConflictedFiles()) ?? []);
      } catch { setMergeInProgress(false); setConflicts([]); }
    } catch (e) {
      flash("err", String(e));
    }
  }, [dir]);

  useEffect(() => { reload(); }, []); // initial

  const run = async (name: string, fn: () => Promise<string | void>, okMsg: string) => {
    setBusy(name);
    try {
      const res = await fn();
      flash("ok", typeof res === "string" && res ? res : okMsg);
      await reload();
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const value: GitmateState = {
    view, setView, dir, setDir,
    status, branches, commits, prs, issues, stashes, checks, mergeInProgress, conflicts,
    toast, busy, setBusy, flash, reload, run,
    service: GitService,
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

// shared tailwind class strings (single source of truth)
export const cls = {
  input: "rounded-md border border-border bg-muted px-3 py-1.5 text-sm outline-none focus:border-[var(--color-ahead)]",
  btn: "rounded-md bg-primary px-3 py-1.5 text-sm font-semibold text-background disabled:opacity-40",
  btnSm: "rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40",
};
