import { useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Tag } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

function locationLabel(t: Tag): { text: string; color: string } {
  if (t.Local && t.Remote) return { text: "both", color: "text-muted-foreground" };
  if (t.Local) return { text: "local only", color: "text-[var(--color-modified)]" };
  if (t.Remote) return { text: "remote only", color: "text-[var(--color-ahead)]" };
  return { text: "", color: "" };
}

export function Tags() {
  const { tags, busy, run, service } = useGit();
  const [name, setName] = useState("");
  const [msg, setMsg] = useState("");
  const [confirmDel, setConfirmDel] = useState<Tag | null>(null);

  const doCreate = () => run("tag-create", async () => {
    await service.CreateTag(name.trim(), msg.trim());
    setName(""); setMsg("");
    return `created ${name.trim()}`;
  }, "tag created");
  const doPush = (n: string) => run(`tag-push-${n}`, async () => {
    await service.PushTag(n); return `pushed ${n} — release workflow may trigger`;
  }, `pushed ${n}`);
  const doSync = () => run("tag-sync", () => service.FetchTags(), "synced tags from origin");
  const doDelete = (t: Tag) => run(`tag-del-${t.Name}`, async () => {
    const where = await service.SmartDeleteTag(t.Name);
    setConfirmDel(null);
    return `deleted ${t.Name} (${where})`;
  }, `deleted ${t.Name}`);

  // what will delete actually remove, for the confirm text
  const delScope = (t: Tag) => (t.Local && t.Remote ? "from local and origin" : t.Local ? "locally" : "from origin");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Tags</h2>
        <button onClick={doSync} disabled={!!busy} className={cls.btnSm} title="fetch tags from origin, prune deleted ones">
          {busy === "tag-sync" ? "…" : "Sync tags"}
        </button>
      </div>

      <div className="space-y-2 rounded-lg border border-border p-3">
        <div className="text-xs text-muted-foreground">Tag the current commit (HEAD). Add a message for an annotated tag.</div>
        <div className="flex gap-2">
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="tag name (e.g. v1.2.0)" className={`${cls.input} flex-1`} />
          <button onClick={doCreate} disabled={!!busy || !name.trim()} className={cls.btn}>{busy === "tag-create" ? "…" : "Create tag"}</button>
        </div>
        <input value={msg} onChange={(e) => setMsg(e.target.value)} placeholder="annotation message (optional)" className={`${cls.input} w-full`} />
      </div>

      <div className="rounded-lg border border-border">
        {(tags ?? []).length ? (tags ?? []).map((t) => {
          const loc = locationLabel(t);
          return (
            <div key={t.Name} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
              <span className="shrink-0 font-semibold text-[var(--color-modified)]">{t.Name}</span>
              <span className={`shrink-0 text-[10px] uppercase tracking-wider ${loc.color}`}>{loc.text}</span>
              <span className="truncate text-xs text-muted-foreground">{t.Subject}</span>
              <span className="ml-auto flex shrink-0 gap-1">
                {t.Local && !t.Remote && (
                  <button onClick={() => doPush(t.Name)} disabled={!!busy} className={cls.btnSm} title="push tag to origin (triggers release)">
                    {busy === `tag-push-${t.Name}` ? "…" : "Push"}
                  </button>
                )}
                <button onClick={() => setConfirmDel(t)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>
                  {busy === `tag-del-${t.Name}` ? "…" : "Delete"}
                </button>
              </span>
            </div>
          );
        }) : <div className="p-3 text-sm italic text-muted-foreground">no tags</div>}
      </div>

      {confirmDel && (
        <ConfirmDialog title="Delete tag?"
          body={<>Delete tag <span className="text-foreground">{confirmDel.Name}</span> {delScope(confirmDel)}.</>}
          confirmLabel="Delete" busy={!!busy}
          onCancel={() => setConfirmDel(null)}
          onConfirm={() => doDelete(confirmDel)} />
      )}
    </div>
  );
}