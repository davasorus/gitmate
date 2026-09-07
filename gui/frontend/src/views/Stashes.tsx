import { useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";

export function Stashes() {
  const { stashes, busy, run, service } = useGit();
  const [msg, setMsg] = useState("");
  const [confirmDrop, setConfirmDrop] = useState<string | null>(null);

  const doSave = () =>
    run(
      "stash-save",
      async () => {
        await service.StashSave(msg.trim(), true);
        setMsg("");
        return "stashed changes";
      },
      "stashed",
    );
  const doPop = (ref: string) =>
    run(`stash-pop-${ref}`, () => service.StashPop(ref), `popped ${ref}`);
  const doApply = (ref: string) =>
    run(`stash-apply-${ref}`, () => service.StashApply(ref), `applied ${ref}`);
  const doDrop = (ref: string) =>
    run(
      `stash-drop-${ref}`,
      async () => {
        await service.StashDrop(ref);
        setConfirmDrop(null);
        return `dropped ${ref}`;
      },
      `dropped ${ref}`,
    );

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div className="flex items-center gap-2">
        <span className="text-[15px] font-semibold tracking-[-0.01em]">Stashes</span>
      </div>

      <div className="flex gap-2">
        <input
          value={msg}
          onChange={(e) => setMsg(e.target.value)}
          placeholder="stash message (optional)"
          className={`${cls.input} flex-1`}
        />
        <button
          onClick={doSave}
          disabled={!!busy}
          className="rounded-lg bg-[var(--color-accent)] px-3.5 py-1.5 text-[12.5px] font-semibold text-white hover:opacity-90 disabled:opacity-40"
        >
          {busy === "stash-save" ? "…" : "Stash changes"}
        </button>
      </div>

      <div className="space-y-1">
        {(stashes ?? []).length ? (
          (stashes ?? []).map((st) => (
            <div
              key={st.Ref}
              className="group flex items-center gap-2.5 rounded-lg border border-border px-3 py-2.5 hover:bg-[var(--color-card)]"
            >
              <span className="shrink-0 font-mono text-[11px] text-[var(--color-modified)]">
                {st.Ref}
              </span>
              <span className="shrink-0 font-mono text-[11px] text-[var(--color-faint)]">
                {st.Branch}
              </span>
              <span className="min-w-0 flex-1 truncate text-[13px]">{st.Message}</span>
              <span className="flex shrink-0 gap-1 opacity-0 group-hover:opacity-100">
                <button
                  onClick={() => doApply(st.Ref)}
                  disabled={!!busy}
                  title="apply but keep the stash"
                  className="rounded-md px-2 py-1 text-[11.5px] text-[var(--color-accent)] hover:bg-[var(--color-muted)]"
                >
                  {busy === `stash-apply-${st.Ref}` ? "…" : "Apply"}
                </button>
                <button
                  onClick={() => doPop(st.Ref)}
                  disabled={!!busy}
                  className="rounded-md px-2 py-1 text-[11.5px] text-[var(--color-accent)] hover:bg-[var(--color-muted)]"
                >
                  {busy === `stash-pop-${st.Ref}` ? "…" : "Pop"}
                </button>
                <button
                  onClick={() => setConfirmDrop(st.Ref)}
                  disabled={!!busy}
                  className="rounded-md px-2 py-1 text-[11.5px] text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10"
                >
                  {busy === `stash-drop-${st.Ref}` ? "…" : "Drop"}
                </button>
              </span>
            </div>
          ))
        ) : (
          <div className="rounded-lg border border-border px-3 py-6 text-center text-[12.5px] italic text-[var(--color-faint)]">
            no stashes
          </div>
        )}
      </div>

      {confirmDrop && (
        <ConfirmDialog
          title="Drop stash?"
          body={
            <>
              Discard <span className="text-foreground">{confirmDrop}</span> without applying it.
              The stashed changes are lost.
            </>
          }
          confirmLabel="Drop"
          busy={!!busy}
          onCancel={() => setConfirmDrop(null)}
          onConfirm={() => doDrop(confirmDrop)}
        />
      )}
    </div>
  );
}
