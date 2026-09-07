import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useRef,
  ReactNode,
} from "react";
import { GitService } from "../bindings/github.com/davasorus/gitmate/gui";
import type {
  Status,
  Commit,
  Branch,
  Stash,
  Tag,
} from "../bindings/github.com/davasorus/gitmate/internal/gitops";
import type { PR, CheckRun, Issue } from "../bindings/github.com/davasorus/gitmate/internal/ghapi";

export type View =
  | "changes"
  | "history"
  | "branches"
  | "prs"
  | "issues"
  | "stashes"
  | "tags"
  | "reflog"
  | "conflicts"
  | "remotes"
  | "labels"
  | "releases"
  | "actions";
type Toast = { kind: "ok" | "err"; msg: string } | null;

export interface GitmateState {
  // view + repo
  view: View;
  setView: (v: View) => void;
  dir: string;
  setDir: (d: string) => void;

  // data
  status: Status | null;
  branches: Branch[];
  commits: Commit[];
  prs: PR[];
  issues: Issue[];
  stashes: Stash[];
  tags: Tag[];
  checks: Record<number, CheckRun[]>;
  mergeInProgress: boolean;
  notRepo: boolean; // current dir is not a git repo → skip all git/GH ops
  undoLabel: string; // pending undoable op label ("" = nothing to undo)
  rebaseInProgress: boolean;
  cherryPickInProgress: boolean;
  revertInProgress: boolean;
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
  const [tags, setTags] = useState<Tag[]>([]);
  const [checks] = useState<Record<number, CheckRun[]>>({});
  const [mergeInProgress, setMergeInProgress] = useState(false);
  const [undoLabel, setUndoLabel] = useState("");
  const [notRepo, setNotRepo] = useState(false);
  const [rebaseInProgress, setRebaseInProgress] = useState(false);
  const [cherryPickInProgress, setCherryPickInProgress] = useState(false);
  const [revertInProgress, setRevertInProgress] = useState(false);
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
      // Guard: in a non-git folder, do NOT fire git/GitHub operations — they
      // cascade-fail (and hammer the API). Clear state and stop here.
      if (!(await GitService.IsRepo())) {
        setNotRepo(true);
        setStatus(null);
        setBranches([]);
        setCommits([]);
        setPRs([]);
        setStashes([]);
        setIssues([]);
        setTags([]);
        setMergeInProgress(false);
        setRebaseInProgress(false);
        setCherryPickInProgress(false);
        setRevertInProgress(false);
        setConflicts([]);
        setUndoLabel("");
        return;
      }
      setNotRepo(false);
      const [s, b, c] = await Promise.all([
        GitService.Status(),
        GitService.Branches(),
        GitService.Log(15),
      ]);
      setStatus(s);
      setBranches(b ?? []);
      setCommits(c ?? []);
      try {
        setPRs((await GitService.PRs("open")) ?? []);
      } catch {
        setPRs([]);
      }
      try {
        setStashes((await GitService.StashList()) ?? []);
      } catch {
        setStashes([]);
      }
      try {
        setIssues((await GitService.Issues("open")) ?? []);
      } catch {
        setIssues([]);
      }
      try {
        setTags((await GitService.ListTags()) ?? []);
      } catch {
        setTags([]);
      }
      try {
        setMergeInProgress(await GitService.MergeInProgress());
        setRebaseInProgress(await GitService.RebaseInProgress());
        const seq = await GitService.SequencerInProgress();
        setCherryPickInProgress(!!seq && !!seq[0]);
        setRevertInProgress(!!seq && !!seq[1]);
        setConflicts((await GitService.ConflictedFiles()) ?? []);
        try {
          const u = await GitService.LastUndoable();
          setUndoLabel(u?.Label ?? "");
        } catch {
          setUndoLabel("");
        }
      } catch {
        setMergeInProgress(false);
        setRebaseInProgress(false);
        setCherryPickInProgress(false);
        setRevertInProgress(false);
        setConflicts([]);
      }
    } catch (e) {
      flash("err", String(e));
    }
  }, [dir]);

  // keep a stable ref to the latest reload so the focus/interval effects don't
  // tear down + recreate every time `dir` (and thus `reload`) changes — that
  // churn was causing the reload cascade.
  const reloadRef = useRef(reload);
  reloadRef.current = reload;
  const notRepoRef = useRef(notRepo);
  notRepoRef.current = notRepo;

  useEffect(() => {
    // load the persisted repo directory (first-class setting) before the first
    // reload; if none is saved, dir stays "" and reload guards via IsRepo.
    (async () => {
      try {
        const saved = await GitService.RepoDir();
        if (saved) setDir(saved);
      } catch {
        /* no saved setting — fine */
      }
      reloadRef.current();
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // initial, once

  // when the repo directory changes (user picked a different repo), reload once.
  // Skips the very first render (the initial effect above handles startup).
  const firstDir = useRef(true);
  useEffect(() => {
    if (firstDir.current) {
      firstDir.current = false;
      return;
    }
    reloadRef.current();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dir]);

  // (1) Reload local state when the window regains focus — reflects changes made
  // in a terminal / another git tool without a manual Reload. Cheap + local (no network).
  useEffect(() => {
    const onFocus = () => {
      reloadRef.current();
    };
    const onVisible = () => {
      if (document.visibilityState === "visible") reloadRef.current();
    };
    window.addEventListener("focus", onFocus);
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      window.removeEventListener("focus", onFocus);
      document.removeEventListener("visibilitychange", onVisible);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // (2) Periodic background fetch (network) so ahead/behind + remote branches stay
  // fresh, then reload. Every 5 minutes. NOT on every reload (fetch is network;
  // reload is cheap+local). Uses fetch --prune so deleted remote branches drop.
  useEffect(() => {
    const id = setInterval(
      async () => {
        if (notRepoRef.current) return; // no repo → nothing to fetch
        try {
          await GitService.Fetch();
          await reloadRef.current();
        } catch {
          /* offline / no remote — non-fatal, try again next tick */
        }
      },
      5 * 60 * 1000,
    );
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
    view,
    setView,
    dir,
    setDir,
    status,
    branches,
    commits,
    prs,
    issues,
    stashes,
    tags,
    checks,
    mergeInProgress,
    notRepo,
    undoLabel,
    rebaseInProgress,
    cherryPickInProgress,
    revertInProgress,
    conflicts,
    toast,
    busy,
    setBusy,
    flash,
    reload,
    run,
    service: GitService,
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

// shared tailwind class strings (single source of truth)
export const cls = {
  input:
    "rounded-md border border-border bg-muted px-3 py-1.5 text-sm outline-none focus:border-[var(--color-ahead)]",
  btn: "rounded-md bg-primary px-3 py-1.5 text-sm font-semibold text-background disabled:opacity-40",
  btnSm: "rounded-md border border-border px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-40",
  // branch-row hierarchy: primary (Switch) stands out; muted (Rename/Delete) recedes
  btnSmPrimary:
    "rounded-md bg-primary px-2 py-0.5 text-xs font-semibold text-background hover:opacity-90 disabled:opacity-40",
  btnSmMuted:
    "rounded-md px-2 py-0.5 text-xs text-muted-foreground hover:bg-muted disabled:opacity-40",
};
