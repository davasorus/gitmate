import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Tag } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function Tags() {
  const { busy, run, service, flash, setBusy } = useGit();
  const [tags, setTags] = useState<Tag[]>([]);
  const [name, setName] = useState("");
  const [msg, setMsg] = useState("");
  const [confirmDel, setConfirmDel] = useState<string | null>(null);

  const load = async () => {
    setBusy("tags-load");
    try { setTags((await service.ListTags()) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => { load(); }, []);

  const doCreate = () => run("tag-create", async () => {
    await service.CreateTag(name.trim(), msg.trim());
    setName(""); setMsg("");
    await load();
    return `created ${name.trim()}`;
  }, "tag created");
  const doDelete = (n: string) => run(`tag-del-${n}`, async () => {
    await service.DeleteTag(n); setConfirmDel(null); await load(); return `deleted ${n}`;
  }, `deleted ${n}`);
  const doPush = (n: string) => run(`tag-push-${n}`, async () => {
    await service.PushTag(n); return `pushed ${n} — release workflow may trigger`;
  }, `pushed ${n}`);

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Tags</h2>

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
          body={<>Delete local tag <span className="text-foreground">{confirmDel}</span>? (Does not delete it from origin.)</>}
          confirmLabel="Delete" busy={!!busy}
          onCancel={() => setConfirmDel(null)} onConfirm={() => doDelete(confirmDel)} />
      )}
    </div>
  );
}