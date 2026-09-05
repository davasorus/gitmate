import { useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";

export function Tags() {
  const { tags, busy, run, service } = useGit();
  const [name, setName] = useState("");
  const [msg, setMsg] = useState("");
  const [confirmDel, setConfirmDel] = useState<string | null>(null);

  const doCreate = () => run("tag-create", async () => {
    await service.CreateTag(name.trim(), msg.trim());
    setName(""); setMsg("");
    return `created ${name.trim()}`;
  }, "tag created");
  const doPush = (n: string) => run(`tag-push-${n}`, async () => {
    await service.PushTag(n); return `pushed ${n} — release workflow may trigger`;
  }, `pushed ${n}`);
  const doSync = () => run("tag-sync", () => service.FetchTags(), "synced tags from origin");
  const doDeleteLocal = (n: string) => run(`tag-del-${n}`, async () => {
    await service.DeleteTag(n); setConfirmDel(null); return `deleted ${n} (local)`;
  }, `deleted ${n}`);
  const doDeleteBoth = (n: string) => run(`tag-del-${n}`, async () => {
    await service.DeleteTag(n);
    await service.DeleteRemoteTag(n);
    setConfirmDel(null);
    return `deleted ${n} (local + origin)`;
  }, `deleted ${n}`);

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
        {(tags ?? []).length ? (tags ?? []).map((t) => (
          <div key={t.Name} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="shrink-0 font-semibold text-[var(--color-modified)]">{t.Name}</span>
            <span className="truncate text-xs text-muted-foreground">{t.Subject}</span>
            <span className="ml-auto flex shrink-0 gap-1">
              <button onClick={() => doPush(t.Name)} disabled={!!busy} className={cls.btnSm} title="push tag to origin (triggers release)">
                {busy === `tag-push-${t.Name}` ? "…" : "Push"}
              </button>
              <button onClick={() => setConfirmDel(t.Name)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>
                {busy === `tag-del-${t.Name}` ? "…" : "Delete"}
              </button>
            </span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no tags</div>}
      </div>

      {confirmDel && (
        <ConfirmDialog title="Delete tag?"
          body={<>Delete tag <span className="text-foreground">{confirmDel}</span>. Choose local only, or local + origin (removes it from GitHub too).</>}
          confirmLabel="Local + origin" busy={!!busy}
          onCancel={() => setConfirmDel(null)}
          onConfirm={() => doDeleteBoth(confirmDel)}
          extra={{ label: "Local only", onClick: () => doDeleteLocal(confirmDel) }} />
      )}
    </div>
  );
}