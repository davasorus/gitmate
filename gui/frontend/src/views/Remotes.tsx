import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Remote } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function Remotes() {
  const { busy, setBusy, flash, run, service } = useGit();
  const [remotes, setRemotes] = useState<Remote[]>([]);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  const [renaming, setRenaming] = useState<{ old: string; next: string } | null>(null);
  const [cloneUrl, setCloneUrl] = useState("");
  const [cloneDest, setCloneDest] = useState("");

  const load = async () => {
    setBusy("remotes-load");
    try { setRemotes((await service.ListRemotes()) ?? []); }
    catch (e) { flash("err", String(e)); } finally { setBusy(""); }
  };
  useEffect(() => { load(); }, []);

  const doAdd = () => run("remote-add", async () => {
    await service.AddRemote(name.trim(), url.trim());
    setName(""); setUrl(""); await load();
    return `added ${name.trim()}`;
  }, "remote added");
  const doRemove = (n: string) => run(`remote-del-${n}`, async () => {
    await service.RemoveRemote(n); setConfirmDel(null); await load();
    return `removed ${n}`;
  }, `removed ${n}`);
  const doRename = () => run("remote-rename", async () => {
    if (!renaming) return;
    await service.RenameRemote(renaming.old, renaming.next.trim());
    setRenaming(null); await load();
    return `renamed ${renaming.old}`;
  }, "remote renamed");

  const doClone = () => run("clone", async () => {
    const path = await service.Clone(cloneUrl.trim(), cloneDest.trim());
    setCloneUrl(""); setCloneDest("");
    return `cloned into ${path} — app now points at it`;
  }, "cloned");

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Remotes</h2>

      <div className="space-y-2 rounded-lg border border-border p-3">
        <div className="text-xs text-muted-foreground">Clone a repository. Leave destination blank to use a folder named after the repo. On success the app switches to the clone.</div>
        <div className="flex flex-wrap gap-2">
          <input value={cloneUrl} onChange={(e) => setCloneUrl(e.target.value)} placeholder="repo URL (https://github.com/owner/repo.git)" className={`${cls.input} flex-1`} />
          <input value={cloneDest} onChange={(e) => setCloneDest(e.target.value)} placeholder="destination folder (optional)" className={`${cls.input} w-64`} />
          <button onClick={doClone} disabled={!!busy || !cloneUrl.trim()} className={cls.btn}>{busy === "clone" ? "…" : "Clone"}</button>
        </div>
      </div>

      <div className="flex flex-wrap gap-2 rounded-lg border border-border p-3">
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="name (e.g. origin)" className={`${cls.input} w-40`} />
        <input value={url} onChange={(e) => setUrl(e.target.value)} placeholder="url (https://github.com/owner/repo.git)" className={`${cls.input} flex-1`} />
        <button onClick={doAdd} disabled={!!busy || !name.trim() || !url.trim()} className={cls.btn}>{busy === "remote-add" ? "…" : "Add remote"}</button>
      </div>

      <div className="rounded-lg border border-border">
        {(remotes ?? []).length ? (remotes ?? []).map((r) => (
          <div key={r.Name} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="shrink-0 font-semibold text-[var(--color-ahead)]">{r.Name}</span>
            <span className="truncate text-xs text-muted-foreground">{r.URL}</span>
            <span className="ml-auto flex shrink-0 gap-1">
              <button onClick={() => setRenaming({ old: r.Name, next: r.Name })} disabled={!!busy} className={cls.btnSm}>Rename</button>
              <button onClick={() => setConfirmDel(r.Name)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>
                {busy === `remote-del-${r.Name}` ? "…" : "Remove"}
              </button>
            </span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no remotes</div>}
      </div>

      {confirmDel && (
        <ConfirmDialog title="Remove remote?"
          body={<>Remove remote <span className="text-foreground">{confirmDel}</span>? This only drops the local reference — commits are untouched, and you can re-add it.</>}
          confirmLabel="Remove" busy={!!busy}
          onCancel={() => setConfirmDel(null)} onConfirm={() => doRemove(confirmDel)} />
      )}
      {renaming && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-96 rounded-lg border border-border bg-card p-4 shadow-lg">
            <div className="mb-2 text-sm font-semibold">Rename remote</div>
            <input autoFocus value={renaming.next} onChange={(e) => setRenaming({ old: renaming.old, next: e.target.value })} className={`${cls.input} mb-4 w-full`} />
            <div className="flex justify-end gap-2">
              <button onClick={() => setRenaming(null)} disabled={!!busy} className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40">Cancel</button>
              <button onClick={doRename} disabled={!!busy || !renaming.next.trim() || renaming.next.trim() === renaming.old} className={cls.btn}>{busy === "remote-rename" ? "…" : "Rename"}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}