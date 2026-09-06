import { useState } from "react";
import type { FileDiff } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

// A pending line comment, keyed by file path + new-file line number.
export type PendingComment = { path: string; line: number; body: string };

// ReviewDiff renders a PR diff like DiffView, but each added/context line on the
// new side is clickable to attach a pending review comment. Pending comments are
// held by the parent (lifted state) and submitted in one batch with the review.
export function ReviewDiff({
  files,
  pending,
  onAdd,
  onRemove,
}: {
  files: FileDiff[];
  pending: PendingComment[];
  onAdd: (c: PendingComment) => void;
  onRemove: (path: string, line: number) => void;
}) {
  const [draft, setDraft] = useState<{ path: string; line: number; body: string } | null>(null);

  if (!files || files.length === 0) {
    return <div className="p-3 text-xs italic text-muted-foreground">no diff</div>;
  }

  const pendingAt = (path: string, line: number) =>
    (pending ?? []).find((c) => c.path === path && c.line === line);

  return (
    <div className="space-y-3">
      {files.map((f, fi) => {
        const path = f.NewPath || f.OldPath;
        return (
          <div key={fi} className="overflow-hidden rounded-md border border-border">
            <div className="border-b border-border bg-muted px-3 py-1.5 text-xs text-[var(--color-ahead)]">{path}</div>
            {f.Binary ? (
              <div className="px-3 py-2 text-xs italic text-muted-foreground">binary file</div>
            ) : (
              (f.Hunks ?? []).map((h, hi) => (
                <div key={hi}>
                  <div className="bg-muted/50 px-3 py-1 text-xs text-muted-foreground">{h.Header}</div>
                  <div className="overflow-x-auto">
                    {(h.Lines ?? []).map((ln, li) => {
                      const add = ln.Kind === "add";
                      const rem = ln.Kind === "remove";
                      const bg = add ? "bg-[var(--color-added)]/10" : rem ? "bg-[var(--color-removed)]/10" : "";
                      const fg = add ? "text-[var(--color-added)]" : rem ? "text-[var(--color-removed)]" : "text-foreground";
                      const marker = add ? "+" : rem ? "-" : " ";
                      // Comment target: any line present on the new side (add or context) with a NewNum.
                      const canComment = !rem && ln.NewNum > 0;
                      const existing = canComment ? pendingAt(path, ln.NewNum) : undefined;
                      return (
                        <div key={li}>
                          <div className={`group flex ${bg} font-mono text-xs leading-5`}>
                            <span className="w-10 shrink-0 select-none px-1 text-right text-muted-foreground/60">{ln.OldNum || ""}</span>
                            <span className="w-10 shrink-0 select-none px-1 text-right text-muted-foreground/60">{ln.NewNum || ""}</span>
                            <span className={`w-4 shrink-0 select-none text-center ${fg}`}>{marker}</span>
                            <span className={`flex-1 whitespace-pre ${fg}`}>{ln.Content}</span>
                            {canComment && !existing && (
                              <button
                                onClick={() => setDraft({ path, line: ln.NewNum, body: "" })}
                                className="mr-1 hidden shrink-0 select-none rounded bg-[var(--color-ahead)]/20 px-1 text-[10px] text-[var(--color-ahead)] group-hover:block"
                                title="comment on this line">＋</button>
                            )}
                          </div>
                          {/* pending comment shown under its line */}
                          {existing && (
                            <div className="ml-24 mr-2 my-1 rounded border border-[var(--color-ahead)]/40 bg-[var(--color-ahead)]/5 px-2 py-1 text-xs">
                              <div className="whitespace-pre-wrap">{existing.body}</div>
                              <button onClick={() => onRemove(path, ln.NewNum)} className="mt-1 text-[10px] text-[var(--color-removed)] hover:underline">remove</button>
                            </div>
                          )}
                          {/* draft editor open on this line */}
                          {draft && draft.path === path && draft.line === ln.NewNum && (
                            <div className="ml-24 mr-2 my-1 space-y-1 rounded border border-border bg-background px-2 py-1">
                              <textarea autoFocus value={draft.body} onChange={(e) => setDraft({ ...draft, body: e.target.value })}
                                        placeholder={`comment on ${path}:${ln.NewNum}`}
                                        className="h-16 w-full resize-y rounded border border-border bg-transparent p-1 text-xs" />
                              <div className="flex gap-1">
                                <button onClick={() => { if (draft.body.trim()) { onAdd({ path, line: ln.NewNum, body: draft.body.trim() }); } setDraft(null); }}
                                        className="rounded border border-border px-2 py-0.5 text-[10px] hover:bg-muted">Add comment</button>
                                <button onClick={() => setDraft(null)} className="rounded border border-border px-2 py-0.5 text-[10px] hover:bg-muted">Cancel</button>
                              </div>
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              ))
            )}
          </div>
        );
      })}
    </div>
  );
}