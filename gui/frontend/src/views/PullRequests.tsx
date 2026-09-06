import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { CheckBadge } from "../components/CheckBadge";
import { ReviewDiff, type PendingComment } from "../components/ReviewDiff";
import type { PR, CheckRun, Review, Reviewer } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";
import type { FileDiff } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

type StateFilter = "open" | "closed" | "all";

export function PullRequests() {
  const { branches, busy, run, service, flash, setBusy } = useGit();
  const [prTitle, setPrTitle] = useState("");
  const [prHead, setPrHead] = useState("");
  const [prBase, setPrBase] = useState("live");
  const [prBody, setPrBody] = useState("");
  const [prResult, setPrResult] = useState("");
  const [checks, setChecks] = useState<Record<number, CheckRun[]>>({});

  // this view owns its PR list + state filter (open/closed/all)
  const [filter, setFilter] = useState<StateFilter>("open");
  const [prs, setPrs] = useState<PR[]>([]);

  const reload = async () => {
    setBusy("prs-load");
    try { setPrs((await service.PRs(filter)) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };

  // fetch on mount and whenever the filter changes — NOT on every render
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setBusy("prs-load");
      try {
        const list = (await service.PRs(filter)) ?? [];
        if (!cancelled) setPrs(list);
      } catch (e) { if (!cancelled) flash("err", String(e)); }
      finally { if (!cancelled) setBusy(""); }
    })();
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter]);

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
    await reload();
    return url;
  }, "PR opened");
  const doMerge = (n: number) => run(`merge-${n}`, async () => { const sha = await service.MergePR(n, "merge"); await reload(); return `merged #${n} (${sha.slice(0, 7)})`; }, `merged #${n}`);
  const doClose = (n: number) => run(`pr-close-${n}`, async () => { await service.SetPRState(n, "closed"); await reload(); return `closed #${n}`; }, `closed #${n}`);
  const doReopen = (n: number) => run(`pr-reopen-${n}`, async () => { await service.SetPRState(n, "open"); await reload(); return `reopened #${n}`; }, `reopened #${n}`);
  const [labelInput, setLabelInput] = useState<Record<number, string>>({});
  const [openReview, setOpenReview] = useState<number | null>(null);
  const [prDiff, setPrDiff] = useState<FileDiff[]>([]);
  const [pending, setPending] = useState<PendingComment[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [reviewers, setReviewers] = useState<Reviewer[]>([]);
  const [reviewBody, setReviewBody] = useState("");
  const [reviewerInput, setReviewerInput] = useState("");

  const loadReview = async (n: number) => {
    setBusy(`review-load-${n}`);
    try {
      setReviews((await service.ListReviews(n)) ?? []);
      setReviewers((await service.ListRequestedReviewers(n)) ?? []);
      setPrDiff((await service.PRDiff(n)) ?? []);
      setPending([]);
      setOpenReview(n);
    } catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  const toggleReview = (n: number) => {
    if (openReview === n) { setOpenReview(null); setReviews([]); setReviewers([]); setReviewBody(""); setPrDiff([]); setPending([]); return; }
    loadReview(n);
  };
  const submitReview = (n: number, event: "APPROVE" | "REQUEST_CHANGES" | "COMMENT") => {
    if (event !== "APPROVE" && !reviewBody.trim()) { flash("err", "A comment is required for request-changes and comment."); return; }
    run(`review-${n}`, async () => {
      await service.SubmitReview(n, event, reviewBody, pending);
      setPending([]);
      setReviewBody("");
      await loadReview(n);
      return `review submitted on #${n}`;
    }, `review submitted on #${n}`);
  };
  const addReviewer = (n: number) => {
    const login = reviewerInput.trim();
    if (!login) return;
    run(`reviewer-add-${n}`, async () => {
      await service.RequestReviewers(n, [login]);
      setReviewerInput("");
      await loadReview(n);
      return `requested ${login}`;
    }, `requested ${login}`);
  };
  const removeReviewer = (n: number, login: string) => run(`reviewer-rm-${n}-${login}`, async () => {
    await service.RemoveReviewer(n, login);
    await loadReview(n);
    return `removed ${login}`;
  }, `removed ${login}`);
  const doAddLabel = (n: number, label: string) => run(`lbl-add-${n}`, async () => { await service.AddLabels(n, [label]); setLabelInput((m) => ({ ...m, [n]: "" })); await reload(); return `labeled #${n}`; }, `labeled #${n}`);
  const doRemoveLabel = (n: number, label: string) => run(`lbl-rm-${n}-${label}`, async () => { await service.RemoveLabel(n, label); await reload(); return `unlabeled #${n}`; }, `unlabeled #${n}`);

  const loadChecks = async (n: number) => {
    setBusy(`checks-${n}`);
    try {
      const runs = (await service.PRChecks(n)) ?? [];
      setChecks((p) => ({ ...p, [n]: runs }));
    } catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Pull requests</h2>
        <div className="flex items-center gap-1 text-xs">
          {(["open", "closed", "all"] as StateFilter[]).map((f) => (
            <button key={f} onClick={() => setFilter(f)}
                    className={`rounded-md border border-border px-2 py-0.5 ${filter === f ? "bg-muted font-semibold" : "hover:bg-muted/60"}`}>
              {f}
            </button>
          ))}
        </div>
      </div>

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
        {busy === "prs-load" ? <div className="p-3 text-sm text-muted-foreground">…</div>
          : (prs ?? []).length ? (prs ?? []).map((p) => (
          <div key={p.Number} className="border-b border-border p-3 last:border-0">
            <div className="flex items-baseline gap-2 text-sm">
              <span className="font-semibold text-[var(--color-ahead)]">#{p.Number}</span>
              <span className="truncate">{p.Title}</span>
              <span className="text-xs text-muted-foreground">@{p.Author}</span>
              <span className="ml-auto flex gap-1">
                <button onClick={() => toggleReview(p.Number)} disabled={!!busy} className={cls.btnSm}>{busy === `review-load-${p.Number}` ? "…" : "Review"}</button>
                <button onClick={() => loadChecks(p.Number)} disabled={!!busy} className={cls.btnSm}>{busy === `checks-${p.Number}` ? "…" : "Checks"}</button>
                <button onClick={() => doMerge(p.Number)} disabled={!!busy} className={cls.btnSm}>{busy === `merge-${p.Number}` ? "…" : "Merge"}</button>
                <button onClick={() => doClose(p.Number)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>{busy === `pr-close-${p.Number}` ? "…" : "Close"}</button>
                <button onClick={() => doReopen(p.Number)} disabled={!!busy} className={cls.btnSm}>{busy === `pr-reopen-${p.Number}` ? "…" : "Reopen"}</button>
              </span>
            </div>
            {checks[p.Number] && (
              <div className="mt-1 space-y-0.5 pl-6">
                {checks[p.Number].length ? checks[p.Number].map((r, i) => <CheckBadge key={i} run={r} />) : <div className="text-xs italic text-muted-foreground">no checks reported</div>}
              </div>
            )}
            {openReview === p.Number && (
              <div className="mt-2 space-y-2 rounded-md border border-border bg-background p-2">
                <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Changed files</div>
                <div className="max-h-96 overflow-auto rounded border border-border">
                  <ReviewDiff files={prDiff ?? []} pending={pending}
                    onAdd={(c) => setPending((ps) => [...ps.filter((x) => !(x.path === c.path && x.line === c.line)), c])}
                    onRemove={(path, line) => setPending((ps) => ps.filter((x) => !(x.path === path && x.line === line)))} />
                </div>
                <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Reviews</div>
                {(reviews ?? []).length ? (reviews ?? []).map((rv) => (
                  <div key={rv.ID} className="flex items-baseline gap-2 text-xs">
                    <span className="font-medium">{rv.Author}</span>
                    <span className={
                      rv.State === "APPROVED" ? "text-[var(--color-added)]"
                      : rv.State === "CHANGES_REQUESTED" ? "text-[var(--color-removed)]"
                      : "text-muted-foreground"}>{rv.State}</span>
                    {rv.Body && <span className="truncate text-muted-foreground">{rv.Body}</span>}
                  </div>
                )) : <div className="text-xs italic text-muted-foreground">no reviews yet</div>}

                <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Requested reviewers</div>
                <div className="flex flex-wrap items-center gap-1">
                  {(reviewers ?? []).map((rr) => (
                    <button key={rr.Login} onClick={() => removeReviewer(p.Number, rr.Login)} disabled={!!busy}
                            className="rounded-full border border-border px-2 py-0.5 text-[10px] hover:bg-[var(--color-removed)]/10" title="click to remove">
                      {rr.Login} ✕
                    </button>
                  ))}
                  <input value={reviewerInput} onChange={(e) => setReviewerInput(e.target.value)}
                         onKeyDown={(e) => { if (e.key === "Enter") addReviewer(p.Number); }}
                         placeholder="+ reviewer" className={`${cls.input} h-6 w-28 px-2 py-0 text-[10px]`} />
                </div>

                <textarea value={reviewBody} onChange={(e) => setReviewBody(e.target.value)}
                          placeholder="review comment (required for request-changes / comment)"
                          className={`${cls.input} h-20 w-full resize-y text-xs`} />
                <div className="flex items-center gap-1">
                  {(pending ?? []).length > 0 && <span className="mr-1 text-[10px] text-[var(--color-ahead)]">{pending.length} pending line comment(s)</span>}
                  <button onClick={() => submitReview(p.Number, "APPROVE")} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-added)]`}>{busy === `review-${p.Number}` ? "…" : "Approve"}</button>
                  <button onClick={() => submitReview(p.Number, "REQUEST_CHANGES")} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)]`}>Request changes</button>
                  <button onClick={() => submitReview(p.Number, "COMMENT")} disabled={!!busy} className={cls.btnSm}>Comment</button>
                </div>
              </div>
            )}
            <div className="mt-1 flex flex-wrap items-center gap-1">
              {(p.Labels ?? []).map((l) => (
                <button key={l} onClick={() => doRemoveLabel(p.Number, l)} disabled={!!busy}
                        className="rounded-full border border-border px-2 py-0.5 text-[10px] hover:bg-[var(--color-removed)]/10" title="click to remove">
                  {l} ✕
                </button>
              ))}
              <input value={labelInput[p.Number] ?? ""} onChange={(e) => setLabelInput((m) => ({ ...m, [p.Number]: e.target.value }))}
                     onKeyDown={(e) => { if (e.key === "Enter" && (labelInput[p.Number] ?? "").trim()) doAddLabel(p.Number, labelInput[p.Number].trim()); }}
                     placeholder="+ label" className={`${cls.input} h-6 w-24 px-2 py-0 text-[10px]`} />
            </div>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no {filter} PRs</div>}
      </div>
    </div>
  );
}