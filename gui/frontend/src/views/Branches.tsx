import { useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";

export function Branches() {
  const { branches, busy, run, service } = useGit();
  const [newBranch, setNewBranch] = useState("");
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  const [renaming, setRenaming] = useState<{ old: string; next: string } | null>(null);

  const doSwitch = (b: string) => run(`switch-${b}`, () => service.Switch(b), `switched to ${b}`);
  const doCreate = () => run("create-branch", async () => { const b = newBranch.trim(); await service.SwitchNew(b); setNewBranch(""); return `created ${b}`; }, "branch created");
  const doDelete = (name: string, force: boolean) => run(`delbranch-${name}`, async () => { await service.DeleteBranch(name, force); setConfirmDel(null); return `deleted ${name}`; }, `deleted ${name}`);
  const doRename = () => run("rename-branch", async () => { if (!renaming) return; const { old, next } = renaming; await service.RenameBranch(old, next.trim()); setRenaming(null); return `renamed ${old} → ${next.trim()}`; }, "branch renamed");
  const doMerge = (b: string) => run(`merge-branch-${b}`, async () => { await service.Merge(b); const c = (await service.ConflictedFiles()) ?? []; return c.length ? `merge hit ${c.length} conflict(s) — see banner` : `merged ${b}`; }, `merged ${b}`);
  const doRebase = (b: string) => run(`rebase-branch-${b}`, async () => { await service.Rebase(b); const c = (await service.ConflictedFiles()) ?? []; return c.length ? `rebase hit ${c.length} conflict(s) — see banner` : `rebased onto ${b}`; }, `rebased onto ${b}`);

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Branches</h2>

      <div className="flex gap-2 rounded-lg border border-border p-3">
        <input value={newBranch} onChange={(e) => setNewBranch(e.target.value)} placeholder="new branch name" className={`${cls.input} flex-1`} />
        <button onClick={doCreate} disabled={!!busy || !newBranch.trim()} className={cls.btn}>{busy === "create-branch" ? "…" : "Create + switch"}</button>
      </div>

      <div className="rounded-lg border border-border">
        {(branches ?? []).map((b) => (
          <div key={b.Name} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            {b.IsCurrent ? <span className="text-[var(--color-added)]">●</span> : <span className="w-2" />}
            <b>{b.Name}</b>
            <span className="text-[var(--color-modified)]">{b.LastHash}</span>
            {!b.Upstream && <span className="text-xs text-muted-foreground">(no upstream)</span>}
            {b.Upstream && (b.Ahead || b.Behind) ? <span className="text-xs text-muted-foreground">↑{b.Ahead} ↓{b.Behind}</span> : null}
            <span className="truncate text-xs text-muted-foreground">{b.LastSubject}</span>
            <span className="ml-auto flex shrink-0 gap-1">
              {!b.IsCurrent && <button onClick={() => doSwitch(b.Name)} disabled={!!busy} className={cls.btnSm}>{busy === `switch-${b.Name}` ? "…" : "Switch"}</button>}
              {!b.IsCurrent && <button onClick={() => doMerge(b.Name)} disabled={!!busy} className={cls.btnSm}>{busy === `merge-branch-${b.Name}` ? "…" : "Merge"}</button>}
              {!b.IsCurrent && <button onClick={() => doRebase(b.Name)} disabled={!!busy} className={cls.btnSm}>{busy === `rebase-branch-${b.Name}` ? "…" : "Rebase"}</button>}
              <button onClick={() => setRenaming({ old: b.Name, next: b.Name })} disabled={!!busy} className={cls.btnSm}>Rename</button>
              {!b.IsCurrent && <button onClick={() => setConfirmDel(b.Name)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>{busy === `delbranch-${b.Name}` ? "…" : "Delete"}</button>}
            </span>
          </div>
        ))}
      </div>

      {confirmDel && (
        <ConfirmDialog title="Delete branch?"
          body={<>Delete branch <span className="text-foreground">{confirmDel}</span>? Safe delete refuses if it has unmerged commits.</>}
          confirmLabel="Force delete" busy={!!busy}
          onCancel={() => setConfirmDel(null)}
          onConfirm={() => doDelete(confirmDel, true)}
          extra={{ label: "Delete", onClick: () => doDelete(confirmDel, false) }} />
      )}
      {renaming && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-96 rounded-lg border border-border bg-card p-4 shadow-lg">
            <div className="mb-2 text-sm font-semibold">Rename branch</div>
            <input autoFocus value={renaming.next} onChange={(e) => setRenaming({ old: renaming.old, next: e.target.value })} className={`${cls.input} mb-4 w-full`} />
            <div className="flex justify-end gap-2">
              <button onClick={() => setRenaming(null)} disabled={!!busy} className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40">Cancel</button>
              <button onClick={doRename} disabled={!!busy || !renaming.next.trim() || renaming.next.trim() === renaming.old} className={cls.btn}>{busy === "rename-branch" ? "…" : "Rename"}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}