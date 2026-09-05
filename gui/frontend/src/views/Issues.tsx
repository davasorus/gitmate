import { useState } from "react";
import { useGit, cls } from "../context";

export function Issues() {
  const { issues, busy, run, service } = useGit();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const doIssue = () => run("issue", async () => { const url = await service.CreateIssue(title.trim(), body); setTitle(""); setBody(""); return url; }, "issue opened");

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Issues</h2>
      <div className="space-y-2 rounded-lg border border-border p-3">
        <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="issue title" className={`${cls.input} w-full`} />
        <textarea value={body} onChange={(e) => setBody(e.target.value)} placeholder="description (optional)" className={`${cls.input} h-24 w-full resize-y`} />
        <button onClick={doIssue} disabled={!!busy || !title.trim()} className={cls.btn}>{busy === "issue" ? "…" : "Open issue"}</button>
      </div>
      <div className="rounded-lg border border-border">
        {(issues ?? []).length ? (issues ?? []).map((i) => (
          <div key={i.Number} className="flex items-baseline gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="font-semibold text-[var(--color-ahead)]">#{i.Number}</span>
            <span className="truncate">{i.Title}</span>
            <span className="ml-auto shrink-0 text-xs text-muted-foreground">@{i.Author}</span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no open issues</div>}
      </div>
    </div>
  );
}
