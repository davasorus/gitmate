import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import type {
  IssueListItem,
  Milestone,
} from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

type StateFilter = "open" | "closed" | "all";

export function Issues() {
  const { busy, run, service, flash, setBusy } = useGit();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");

  const [filter, setFilter] = useState<StateFilter>("open");
  const [issues, setIssues] = useState<IssueListItem[]>([]);
  const [milestones, setMilestones] = useState<Milestone[]>([]);
  const [sel, setSel] = useState<number | null>(null);
  const [showNew, setShowNew] = useState(false);

  const reload = async () => {
    setBusy("issues-load");
    try {
      setIssues((await service.IssuesRich(filter)) ?? []);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  // fetch on mount and whenever the filter changes — NOT on every render
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setBusy("issues-load");
      try {
        const list = (await service.IssuesRich(filter)) ?? [];
        if (!cancelled) setIssues(list);
        try {
          const ms = (await service.ListMilestones("open")) ?? [];
          if (!cancelled) setMilestones(ms);
        } catch {
          /* milestones optional */
        }
      } catch (e) {
        if (!cancelled) flash("err", String(e));
      } finally {
        if (!cancelled) setBusy("");
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter]);

  const doIssue = () =>
    run(
      "issue",
      async () => {
        const url = await service.CreateIssue(title.trim(), body);
        setTitle("");
        setBody("");
        await reload();
        return url;
      },
      "issue opened",
    );
  const doClose = (n: number) =>
    run(
      `issue-close-${n}`,
      async () => {
        await service.SetIssueState(n, "closed");
        await reload();
        return `closed #${n}`;
      },
      `closed #${n}`,
    );
  const doReopen = (n: number) =>
    run(
      `issue-reopen-${n}`,
      async () => {
        await service.SetIssueState(n, "open");
        await reload();
        return `reopened #${n}`;
      },
      `reopened #${n}`,
    );
  const [labelInput, setLabelInput] = useState<Record<number, string>>({});
  const [assigneeInput, setAssigneeInput] = useState<Record<number, string>>({});
  const doAddAssignee = (n: number, user: string) =>
    run(
      `assign-${n}`,
      async () => {
        await service.AddAssignees(n, [user]);
        setAssigneeInput((m) => ({ ...m, [n]: "" }));
        return `assigned @${user}`;
      },
      "assigned",
    );
  const doRemoveAssignee = (n: number, user: string) =>
    run(
      `unassign-${n}-${user}`,
      async () => {
        await service.RemoveAssignees(n, [user]);
        return `unassigned @${user}`;
      },
      "unassigned",
    );
  const doLock = (n: number) =>
    run(
      `lock-${n}`,
      async () => {
        await service.LockConversation(n, "");
        return `locked #${n}`;
      },
      "locked",
    );
  const doUnlock = (n: number) =>
    run(
      `unlock-${n}`,
      async () => {
        await service.UnlockConversation(n);
        return `unlocked #${n}`;
      },
      "unlocked",
    );
  const doSetMilestone = (n: number, milestone: number) =>
    run(
      `milestone-${n}`,
      async () => {
        await service.SetMilestone(n, milestone);
        return milestone === 0 ? `cleared milestone on #${n}` : `set milestone on #${n}`;
      },
      "milestone set",
    );
  const doAddLabel = (n: number, label: string) =>
    run(
      `lbl-add-${n}`,
      async () => {
        await service.AddLabels(n, [label]);
        setLabelInput((m) => ({ ...m, [n]: "" }));
        await reload();
        return `labeled #${n}`;
      },
      `labeled #${n}`,
    );
  const doRemoveLabel = (n: number, label: string) =>
    run(
      `lbl-rm-${n}-${label}`,
      async () => {
        await service.RemoveLabel(n, label);
        await reload();
        return `unlabeled #${n}`;
      },
      `unlabeled #${n}`,
    );


  const selected = (issues ?? []).find((x) => x.Number === sel) ?? null;

  return (
    <div className="flex h-[calc(100vh-116px)] gap-0">
      {/* LEFT: issue list */}
      <div className="flex w-[340px] shrink-0 flex-col border-r border-border pr-3">
        <div className="mb-2 flex items-center gap-2 px-1">
          <div className="flex items-center gap-0.5 rounded-lg border border-border bg-[var(--color-muted)] p-0.5">
            {(["open", "closed", "all"] as StateFilter[]).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`rounded-md px-2.5 py-1 text-[11.5px] font-medium ${filter === f ? "bg-[var(--color-raised)] text-[var(--color-foreground)]" : "text-[var(--color-faint)]"}`}
              >
                {f}
              </button>
            ))}
          </div>
          <button
            onClick={() => setShowNew((v) => !v)}
            className="ml-auto text-[11.5px] font-semibold text-[var(--color-accent)] hover:underline"
          >
            {showNew ? "Cancel" : "+ New issue"}
          </button>
        </div>

        {showNew && (
          <div className="mb-2 space-y-2 rounded-lg border border-border p-2">
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="issue title"
              className={`${cls.input} w-full py-1 text-xs`}
            />
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="description (optional)"
              className={`${cls.input} h-20 w-full resize-y text-xs`}
            />
            <button
              onClick={doIssue}
              disabled={!!busy || !title.trim()}
              className="w-full rounded-lg bg-[var(--color-accent)] px-3 py-1.5 text-[12px] font-semibold text-white disabled:opacity-40"
            >
              {busy === "issue" ? "…" : "Open issue"}
            </button>
          </div>
        )}

        <div className="min-h-0 flex-1 overflow-y-auto">
          {busy === "issues-load" ? (
            <div className="px-2.5 py-2 text-[12px] text-[var(--color-faint)]">…</div>
          ) : (issues ?? []).length ? (
            (issues ?? []).map((i) => (
              <div
                key={i.Number}
                onClick={() => setSel(i.Number)}
                className={`cursor-pointer rounded-lg px-2.5 py-2 ${sel === i.Number ? "bg-[var(--color-accent-dim)]" : "hover:bg-[var(--color-card)]"}`}
              >
                <div className="flex items-center gap-2">
                  <span className="shrink-0 font-mono text-[11px] text-[var(--color-faint)]">
                    #{i.Number}
                  </span>
                  <span className="min-w-0 flex-1 truncate text-[13px] font-medium">
                    {i.Title}
                  </span>
                </div>
                <div className="mt-1 flex items-center gap-1.5 pl-1">
                  <span className="text-[11px] text-[var(--color-faint)]">@{i.Author}</span>
                  {(i.Labels ?? []).slice(0, 3).map((l) => (
                    <span
                      key={l}
                      className="rounded-full border border-border px-1.5 py-0.5 text-[9.5px] text-[var(--color-muted-foreground)]"
                    >
                      {l}
                    </span>
                  ))}
                </div>
              </div>
            ))
          ) : (
            <div className="px-2.5 py-2 text-[12px] italic text-[var(--color-faint)]">
              no {filter} issues
            </div>
          )}
        </div>
      </div>

      {/* RIGHT: issue detail */}
      <div className="min-w-0 flex-1 overflow-y-auto pl-4">
        {!selected ? (
          <div className="grid h-full place-items-center text-[13px] text-[var(--color-faint)]">
            Select an issue
          </div>
        ) : (
          [selected].map((i) => (
            <div key={i.Number} className="space-y-4">
              <div className="border-b border-border pb-3">
                <div className="flex items-baseline gap-2">
                  <span className="font-mono text-[12px] text-[var(--color-faint)]">
                    #{i.Number}
                  </span>
                  <span className="text-[16px] font-semibold tracking-[-0.02em]">{i.Title}</span>
                </div>
                <div className="mt-1 text-[12px] text-[var(--color-muted-foreground)]">
                  @{i.Author}
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <button
                    onClick={() => doClose(i.Number)}
                    disabled={!!busy}
                    className="rounded-lg border border-border px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10 disabled:opacity-40"
                  >
                    {busy === `issue-close-${i.Number}` ? "…" : "Close"}
                  </button>
                  <button
                    onClick={() => doReopen(i.Number)}
                    disabled={!!busy}
                    className="rounded-lg border border-border px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
                  >
                    {busy === `issue-reopen-${i.Number}` ? "…" : "Reopen"}
                  </button>
                  <button
                    onClick={() => doLock(i.Number)}
                    disabled={!!busy}
                    className="rounded-lg border border-border px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
                  >
                    {busy === `lock-${i.Number}` ? "…" : "Lock"}
                  </button>
                  <button
                    onClick={() => doUnlock(i.Number)}
                    disabled={!!busy}
                    className="rounded-lg border border-border px-3.5 py-1.5 text-[12.5px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
                  >
                    {busy === `unlock-${i.Number}` ? "…" : "Unlock"}
                  </button>
                </div>
              </div>

              <div>
                <div className="mb-2 text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
                  ASSIGNEES
                </div>
                <div className="flex flex-wrap items-center gap-1.5">
                  {(i.Assignees ?? []).map((a) => (
                    <button
                      key={a}
                      onClick={() => doRemoveAssignee(i.Number, a)}
                      disabled={!!busy}
                      className="rounded-full border border-border px-2 py-0.5 text-[11px] text-[var(--color-muted-foreground)] hover:bg-[var(--color-removed)]/10"
                    >
                      @{a} ✕
                    </button>
                  ))}
                  <input
                    value={assigneeInput[i.Number] ?? ""}
                    onChange={(e) =>
                      setAssigneeInput((m) => ({ ...m, [i.Number]: e.target.value }))
                    }
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && (assigneeInput[i.Number] ?? "").trim())
                        doAddAssignee(i.Number, assigneeInput[i.Number].trim());
                    }}
                    placeholder="+ assignee"
                    className={`${cls.input} h-7 w-32 px-2 py-0 text-[11px]`}
                  />
                </div>
              </div>

              <div>
                <div className="mb-2 text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
                  LABELS
                </div>
                <div className="flex flex-wrap items-center gap-1.5">
                  {(i.Labels ?? []).map((l) => (
                    <button
                      key={l}
                      onClick={() => doRemoveLabel(i.Number, l)}
                      disabled={!!busy}
                      className="rounded-full border border-border px-2 py-0.5 text-[11px] hover:bg-[var(--color-removed)]/10"
                    >
                      {l} ✕
                    </button>
                  ))}
                  <input
                    value={labelInput[i.Number] ?? ""}
                    onChange={(e) => setLabelInput((m) => ({ ...m, [i.Number]: e.target.value }))}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && (labelInput[i.Number] ?? "").trim())
                        doAddLabel(i.Number, labelInput[i.Number].trim());
                    }}
                    placeholder="+ label"
                    className={`${cls.input} h-7 w-28 px-2 py-0 text-[11px]`}
                  />
                </div>
              </div>

              {milestones.length > 0 && (
                <div>
                  <div className="mb-2 text-[11px] font-semibold tracking-wide text-[var(--color-faint)]">
                    MILESTONE
                  </div>
                  <select
                    disabled={!!busy}
                    defaultValue=""
                    onChange={(e) => {
                      const v = parseInt(e.target.value, 10);
                      if (!Number.isNaN(v)) doSetMilestone(i.Number, v);
                    }}
                    className={`${cls.input} h-8 px-2 py-0 text-[12px]`}
                  >
                    <option value="">milestone…</option>
                    {milestones.map((m) => (
                      <option key={m.Number} value={m.Number}>
                        {m.Title}
                      </option>
                    ))}
                    <option value="0">(clear)</option>
                  </select>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}