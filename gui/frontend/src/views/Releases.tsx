import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Release, Asset } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

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
  const [openAssets, setOpenAssets] = useState<number | null>(null);
  const [assets, setAssets] = useState<Asset[]>([]);

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

  const loadAssets = async (releaseID: number) => {
    setBusy(`assets-${releaseID}`);
    try { setAssets((await service.ListAssets(releaseID)) ?? []); setOpenAssets(releaseID); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  const toggleAssets = (releaseID: number) => {
    if (openAssets === releaseID) { setOpenAssets(null); setAssets([]); return; }
    loadAssets(releaseID);
  };

  // --- Upload: immutable releases reject; every failure surfaces as an error toast ---
  const onUpload = (r: Release, file: File) => {
    if (r.Immutable) {
      flash("err", `Release ${r.TagName} is immutable — assets cannot be added.`);
      return;
    }
    run(`asset-up-${r.ID}`, async () => {
      const buf = await file.arrayBuffer();
      let binary = "";
      const bytes = new Uint8Array(buf);
      const chunk = 0x8000;
      for (let i = 0; i < bytes.length; i += chunk) {
        binary += String.fromCharCode.apply(null, Array.from(bytes.subarray(i, i + chunk)));
      }
      const b64 = btoa(binary);
      await service.UploadAsset(r.ID, file.name, b64); // throws → run() catches → error toast
      await loadAssets(r.ID);
      return `uploaded ${file.name}`;
    }, `uploaded ${file.name}`);
  };

  const onDownload = (a: Asset) => run(`asset-dl-${a.ID}`, async () => {
    const b64 = await service.DownloadAsset(a.ID);
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    const url = URL.createObjectURL(new Blob([bytes]));
    const link = document.createElement("a");
    link.href = url; link.download = a.Name;
    document.body.appendChild(link); link.click(); link.remove();
    URL.revokeObjectURL(url);
    return `downloaded ${a.Name}`;
  }, `downloaded ${a.Name}`);

  const onDeleteAsset = (r: Release, a: Asset) => {
    if (r.Immutable) {
      flash("err", `Release ${r.TagName} is immutable — assets cannot be deleted.`);
      return;
    }
    run(`asset-del-${a.ID}`, async () => {
      await service.DeleteAsset(a.ID);
      await loadAssets(r.ID);
      return `deleted ${a.Name}`;
    }, `deleted ${a.Name}`);
  };

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

  const doDelete = (r: Release) => {
    if (r.Immutable) {
      flash("err", `Release ${r.TagName} is immutable — it cannot be deleted.`);
      setConfirmDel(null);
      return;
    }
    run(`release-del-${r.ID}`, async () => {
      await service.DeleteRelease(r.ID); setConfirmDel(null); await reload();
      return `deleted ${r.TagName}`;
    }, `deleted ${r.TagName}`);
  };

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
          <div key={r.ID} className="border-b border-border last:border-0">
            <div className="flex items-center gap-2 px-3 py-1.5 text-sm">
              <span className="shrink-0 font-semibold text-[var(--color-modified)]">{r.TagName}</span>
              <span className="truncate">{r.Name}</span>
              {r.Draft && <span className="shrink-0 rounded bg-border px-1 text-[10px] uppercase">draft</span>}
              {r.Prerelease && <span className="shrink-0 rounded bg-[var(--color-behind)]/20 px-1 text-[10px] uppercase text-[var(--color-behind)]">pre</span>}
              {r.Immutable && <span className="shrink-0 rounded bg-[var(--color-ahead)]/20 px-1 text-[10px] uppercase text-[var(--color-ahead)]" title="immutable: assets and this release cannot be added, changed, or deleted">🔒 immutable</span>}
              <span className="ml-auto flex shrink-0 gap-1">
                <button onClick={() => toggleAssets(r.ID)} disabled={!!busy} className={cls.btnSm}>{busy === `assets-${r.ID}` ? "…" : "Assets"}</button>
                <a href={r.URL} target="_blank" rel="noreferrer" className={cls.btnSm}>View</a>
                <button onClick={() => setEditing(r)} disabled={!!busy} className={cls.btnSm}>Edit</button>
                <button onClick={() => setConfirmDel(r)} disabled={!!busy || r.Immutable}
                        title={r.Immutable ? "immutable release — cannot delete" : ""}
                        className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10 disabled:opacity-40 disabled:cursor-not-allowed`}>
                  {busy === `release-del-${r.ID}` ? "…" : "Delete"}
                </button>
              </span>
            </div>
            {openAssets === r.ID && (
              <div className="space-y-1 bg-background px-3 pb-2 pt-1">
                {(assets ?? []).length ? (assets ?? []).map((a) => (
                  <div key={a.ID} className="flex items-center gap-2 text-xs">
                    <span className="truncate">{a.Name}</span>
                    <span className="shrink-0 text-muted-foreground">{(a.Size / 1024).toFixed(1)} KB</span>
                    <span className="ml-auto flex shrink-0 gap-1">
                      <button onClick={() => onDownload(a)} disabled={!!busy} className={cls.btnSm}>{busy === `asset-dl-${a.ID}` ? "…" : "Download"}</button>
                      <button onClick={() => onDeleteAsset(r, a)} disabled={!!busy || r.Immutable}
                              title={r.Immutable ? "immutable release — cannot delete asset" : ""}
                              className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10 disabled:opacity-40 disabled:cursor-not-allowed`}>
                        {busy === `asset-del-${a.ID}` ? "…" : "Delete"}
                      </button>
                    </span>
                  </div>
                )) : <div className="text-xs italic text-muted-foreground">no assets</div>}
                {r.Immutable ? (
                  <div className="text-[11px] italic text-muted-foreground">🔒 This release is immutable — assets cannot be added or removed.</div>
                ) : (
                  <label className={`${cls.btnSm} inline-block cursor-pointer`}>
                    {busy === `asset-up-${r.ID}` ? "uploading…" : "Upload file"}
                    <input type="file" className="hidden" disabled={!!busy}
                           onChange={(e) => { const f = e.target.files?.[0]; if (f) onUpload(r, f); e.target.value = ""; }} />
                  </label>
                )}
              </div>
            )}
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