import { useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";

export function Stashes() {
  const { stashes, busy, run, service } = useGit();
  const [msg, setMsg] = useState("");
  const [confirmDrop, setConfirmDrop] = useState<string | null>(null);

  const doSave = () => run("stash-save", async () => { await service.StashSave(msg.trim(), true); setMsg(""); return "stashed changes"; }, "stashed");
  const doPop = (ref: string) => run(`stash-pop-${ref}`, () => service.StashPop(ref), `popped ${ref}`);
  const doDrop = (ref: string) => run(`stash-drop-${ref}`, async () => { await service.StashDrop(ref); setConfirmDrop(null); return `dropped ${ref}`; }, `dropped ${ref}`);

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Stashes</h2>
      <div className="flex gap-2 rounded-lg border border-border p-3">
        <input value={msg} onChange={(e) => setMsg(e.target.value)} placeholder="stash message (optional)" className={`${cls.input} flex-1`} />
        <button onClick={doSave} disabled={!!busy} className={cls.btn}>{busy === "stash-save" ? "…" : "Stash changes"}</button>
      </div>
      <div className="rounded-lg border border-border">
        {(stashes ?? []).length ? (stashes ?? []).map((st) => (
          <div key={st.Ref} className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0">
            <span className="shrink-0 text-[var(--color-modified)]">{st.Ref}</span>
            <span className="shrink-0 text-xs text-muted-foreground">{st.Branch}</span>
            <span className="truncate">{st.Message}</span>
            <span className="ml-auto flex shrink-0 gap-1">
              <button onClick={() => doPop(st.Ref)} disabled={!!busy} className={cls.btnSm}>{busy === `stash-pop-${st.Ref}` ? "…" : "Pop"}</button>
              <button onClick={() => setConfirmDrop(st.Ref)} disabled={!!busy} className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}>{busy === `stash-drop-${st.Ref}` ? "…" : "Drop"}</button>
            </span>
          </div>
        )) : <div className="p-3 text-sm italic text-muted-foreground">no stashes</div>}
      </div>
      {confirmDrop && (
        <ConfirmDialog title="Drop stash?"
          body={<>Discard <span className="text-foreground">{confirmDrop}</span> without applying it. The stashed changes are lost.</>}
          confirmLabel="Drop" busy={!!busy}
          onCancel={() => setConfirmDrop(null)} onConfirm={() => doDrop(confirmDrop)} />
      )}
    </div>
  );
}
