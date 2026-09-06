import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Release } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

export function Releases() {
  const { busy, run, service, flash, setBusy } = useGit();
  const [releases, setReleases] = useState<Release[]>([]);
  const [tag, setTag] = useState("");
  const [name, setName] = useState("");
  const [body, setBody] = useState("");
  const [draft, setDraft] = useState(false);
  const [prerelease, setPrerelease] = useState(false);
  const [confirmDel, setConfirmDel] = useState<Release | null>(null);
  const [editing, setEditing] = useState<Release | null>(null);

  const reload = async () => {
    setBusy("releases-load");
    try { setReleases((await service.ListReleases()) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setBusy("releases-load");
      try { const r = (await service.ListReleases()) ?? []; if (!cancelled) setReleases(r); }
      catch (e) { if (!cancelled) flash("err", String(e)); }
      finally { if (!cancelled) setBusy(""); }
    })();
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const doGenNotes = () => run("gen-notes", async () => {
    if (!tag.trim()) return;
    const nb = await service.GenerateReleaseNotes(tag.trim());
    if (nb && nb.length === 2) { setName((n) => n || nb[0]); setBody(nb[1]); }
    return "notes generated";
  }, "notes generated");

  const doCreate = () => run("release-create", async () => {
    await service.CreateRelease(tag.trim(), name.trim(), body, draft, prerelease);
    setTag(""); setName(""); setBody(""); setDraft(false); setPrerelease(false);
    await reload();
    return `released ${tag.trim()}`;
  }, "release created");

  const doDelete = (r: Release) => run(`release-del-${r.ID}`, async () => {
    await service.DeleteRelease(r.ID); setConfirmDel(null); await reload();
    return `deleted ${r.TagName}`;
  }, `deleted ${r.TagName}`);

  const doEdit = () => run("release-edit", async () => {
    if (!editing) return;
    await service.EditRelease(editing.ID, editing.Name, editing.Body, editing.Draft, editing.Prerelease);
    setEditing(null); await reload();
    return `edited ${editing.TagName}`;
  }, "release edited");

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Releases</h2>

      <div className="space-y-2 rounded-lg border border-border p-3">
        <div className="text-xs text-muted-foreground">Create a release on a tag. Use Generate notes to auto-fill from merged PRs/commits.</div>
        <div className="flex flex-wrap gap-2">
          <input value={tag} onChange={(e) => setTag(e.target.value)} placeholder="tag (e.g. v1.2.0)" className={`${cls.input} w-40`} />
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="title (optional)" className={`${cls.input} flex-1`} />
          <button onClick={doGenNotes} disabled={!!busy || !tag.trim()} className={cls.btnSm}>{busy === "gen-notes" ? "…" : "Generate notes"}</button>
        </div>
        <textarea value={body} onChange={(e) => setBody(e.target.value)} placeholder="release notes" className={`${cls.input} h-28 w-full resize-y`} />
        <div className="flex items-center gap-4 text-xs">
          <label className="flex items-center gap-1"><input type="checkbox" checked={draft} onChange={(e) => setDraft(e.target.checked)} /> draft</label>
          <label className="flex items-center gap-1"><input type="checkbox" checked={prerelease} onChange={(e) => setPrerelease(e.target.checked)} /> prerelease</label>
          <button onClick={doCreate} disabled={!!busy || !tag.trim()} className={`${cls.btn} ml-auto`}>{busy === "release-create" ? "…" : "Create release"}</button>
        </div>
      </div>

      <div className="rounded-lg border border-border">
        {busy === "releases-load" ? <div className="p-3 text-sm text-muted-foreground">…</div>
          : (releases ?? []).length ? (releases ?? []).map((r) => (
          <div key={r.ID} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="shrink-0 font-semibold text-[var(--color-modified)]">{r.TagName}</span>
            <span className="truncate">{r.Name}</span>
            {r.Draft && <span className="shrink-0 rounded bg-border px-1 text-[10px] uppercase">draft</span>}
            {r.Prerelease && <span className="shrink-0 rounded bg-[var(--color-behind)]/20 px-1 text-[10px] uppercase text-[var(--color-behind)]">pre</span>}
            <span className="ml-auto flex shrink-0 gap-1">
              <a href={r.URL} target="_blank" rel="noreferrer" className={cls.btnSm}>View</a>
              <button onClick={() => setEditing(r)} disabled={!!busy} className={cls.btnSm}>Edit</button>
              <button onClick={() => setConfirmDel(r)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>{busy === `release-del-${r.ID}` ? "…" : "Delete"}</button>
            </span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no releases</div>}
      </div>

      {confirmDel && (
        <ConfirmDialog title="Delete release?"
          body={<>Delete the release for <span className="text-foreground">{confirmDel.TagName}</span>? The underlying tag is NOT deleted — only the release.</>}
          confirmLabel="Delete" busy={!!busy}
          onCancel={() => setConfirmDel(null)} onConfirm={() => doDelete(confirmDel)} />
      )}
      {editing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-[32rem] space-y-3 rounded-lg border border-border bg-card p-4 shadow-lg">
            <div className="text-sm font-semibold">Edit release — {editing.TagName}</div>
            <input value={editing.Name} onChange={(e) => setEditing({ ...editing, Name: e.target.value })} placeholder="title" className={`${cls.input} w-full`} />
            <textarea value={editing.Body} onChange={(e) => setEditing({ ...editing, Body: e.target.value })} placeholder="notes" className={`${cls.input} h-40 w-full resize-y`} />
            <div className="flex items-center gap-4 text-xs">
              <label className="flex items-center gap-1"><input type="checkbox" checked={editing.Draft} onChange={(e) => setEditing({ ...editing, Draft: e.target.checked })} /> draft</label>
              <label className="flex items-center gap-1"><input type="checkbox" checked={editing.Prerelease} onChange={(e) => setEditing({ ...editing, Prerelease: e.target.checked })} /> prerelease</label>
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setEditing(null)} disabled={!!busy} className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40">Cancel</button>
              <button onClick={doEdit} disabled={!!busy} className={cls.btn}>{busy === "release-edit" ? "…" : "Save"}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}