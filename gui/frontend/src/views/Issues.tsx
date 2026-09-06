import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import type { Issue } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

type StateFilter = "open" | "closed" | "all";

export function Issues() {
  const { busy, run, service, flash, setBusy } = useGit();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");

  const [filter, setFilter] = useState<StateFilter>("open");
  const [issues, setIssues] = useState<Issue[]>([]);

  const reload = async () => {
    setBusy("issues-load");
    try {
      setIssues((await service.Issues(filter)) ?? []);
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
        const list = (await service.Issues(filter)) ?? [];
        if (!cancelled) setIssues(list);
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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
          Issues
        </h2>
        <div className="flex items-center gap-1 text-xs">
          {(["open", "closed", "all"] as StateFilter[]).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`rounded-md border border-border px-2 py-0.5 ${filter === f ? "bg-muted font-semibold" : "hover:bg-muted/60"}`}
            >
              {f}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-2 rounded-lg border border-border p-3">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="issue title"
          className={`${cls.input} w-full`}
        />
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="description (optional)"
          className={`${cls.input} h-24 w-full resize-y`}
        />
        <button onClick={doIssue} disabled={!!busy || !title.trim()} className={cls.btn}>
          {busy === "issue" ? "…" : "Open issue"}
        </button>
      </div>

      <div className="rounded-lg border border-border">
        {busy === "issues-load" ? (
          <div className="p-3 text-sm text-muted-foreground">…</div>
        ) : (issues ?? []).length ? (
          (issues ?? []).map((i) => (
            <div key={i.Number} className="border-b border-border px-3 py-1.5 last:border-0">
              <div className="flex items-center gap-2 text-sm">
                <span className="shrink-0 font-semibold text-[var(--color-ahead)]">
                  #{i.Number}
                </span>
                <span className="truncate">{i.Title}</span>
                <span className="text-xs text-muted-foreground">@{i.Author}</span>
                <span className="ml-auto flex shrink-0 gap-1">
                  <button
                    onClick={() => doClose(i.Number)}
                    disabled={!!busy}
                    className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}
                  >
                    {busy === `issue-close-${i.Number}` ? "…" : "Close"}
                  </button>
                  <button
                    onClick={() => doReopen(i.Number)}
                    disabled={!!busy}
                    className={cls.btnSm}
                  >
                    {busy === `issue-reopen-${i.Number}` ? "…" : "Reopen"}
                  </button>
                </span>
              </div>
              <div className="mt-1 flex flex-wrap items-center gap-1">
                {(i.Labels ?? []).map((l) => (
                  <button
                    key={l}
                    onClick={() => doRemoveLabel(i.Number, l)}
                    disabled={!!busy}
                    className="rounded-full border border-border px-2 py-0.5 text-[10px] hover:bg-[var(--color-removed)]/10"
                    title="click to remove"
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
                  className={`${cls.input} h-6 w-24 px-2 py-0 text-[10px]`}
                />
              </div>
            </div>
          ))
        ) : (
          <div className="p-3 text-sm italic text-muted-foreground">no {filter} issues</div>
        )}
      </div>
    </div>
  );
}
