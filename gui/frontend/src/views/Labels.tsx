import { useEffect, useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Label } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

const randomHex = () =>
  Math.floor(Math.random() * 0xffffff)
    .toString(16)
    .padStart(6, "0");
const normHex = (h: string) => h.replace(/^#/, "").toLowerCase();

export function Labels() {
  const { busy, run, service, flash, setBusy } = useGit();
  const [labels, setLabels] = useState<Label[]>([]);
  const [name, setName] = useState("");
  const [color, setColor] = useState(randomHex());
  const [desc, setDesc] = useState("");
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  const [editing, setEditing] = useState<(Label & { newName: string }) | null>(null);

  const reload = async () => {
    setBusy("labels-load");
    try {
      setLabels((await service.ListLabels()) ?? []);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setBusy("labels-load");
      try {
        const l = (await service.ListLabels()) ?? [];
        if (!cancelled) setLabels(l);
      } catch (e) {
        if (!cancelled) flash("err", String(e));
      } finally {
        if (!cancelled) setBusy("");
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const doCreate = () =>
    run(
      "label-create",
      async () => {
        await service.CreateLabel(name.trim(), color.trim().replace(/^#/, ""), desc.trim());
        setName("");
        setColor("cccccc");
        setDesc("");
        await reload();
        return `created ${name.trim()}`;
      },
      "label created",
    );
  const doDelete = (n: string) =>
    run(
      `label-del-${n}`,
      async () => {
        await service.DeleteLabel(n);
        setConfirmDel(null);
        await reload();
        return `deleted ${n}`;
      },
      `deleted ${n}`,
    );
  const doEdit = () =>
    run(
      "label-edit",
      async () => {
        if (!editing) return;
        await service.EditLabel(
          editing.Name,
          editing.newName.trim() || editing.Name,
          editing.Color.replace(/^#/, ""),
          editing.Description,
        );
        setEditing(null);
        await reload();
        return `edited ${editing.Name}`;
      },
      "label edited",
    );

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
        Labels
      </h2>

      <div className="flex flex-wrap items-center gap-2 rounded-lg border border-border p-3">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="label name"
          className={`${cls.input} w-40`}
        />
        <input
          type="color"
          value={`#${normHex(color)}`}
          onChange={(e) => setColor(normHex(e.target.value))}
          className="h-8 w-10 shrink-0 cursor-pointer rounded border border-border bg-transparent p-0"
          title="pick a color"
        />
        <button
          type="button"
          onClick={() => setColor(randomHex())}
          className={cls.btnSm}
          title="random color"
        >
          🎲
        </button>
        <input
          value={color}
          onChange={(e) => setColor(normHex(e.target.value))}
          placeholder="hex"
          className={`${cls.input} w-24`}
        />
        <input
          value={desc}
          onChange={(e) => setDesc(e.target.value)}
          placeholder="description (optional)"
          className={`${cls.input} flex-1`}
        />
        <button onClick={doCreate} disabled={!!busy || !name.trim()} className={cls.btn}>
          {busy === "label-create" ? "…" : "Create"}
        </button>
      </div>

      <div className="rounded-lg border border-border">
        {busy === "labels-load" ? (
          <div className="p-3 text-sm text-muted-foreground">…</div>
        ) : (labels ?? []).length ? (
          (labels ?? []).map((l) => (
            <div
              key={l.Name}
              className="flex items-center gap-2 border-b border-border px-3 py-1.5 text-sm last:border-0"
            >
              <span
                className="h-4 w-4 shrink-0 rounded-full border border-border"
                style={{ backgroundColor: `#${l.Color}` }}
              />
              <span className="shrink-0 font-medium">{l.Name}</span>
              <span className="truncate text-xs text-muted-foreground">{l.Description}</span>
              <span className="ml-auto flex shrink-0 gap-1">
                <button
                  onClick={() => setEditing({ ...l, newName: l.Name })}
                  disabled={!!busy}
                  className={cls.btnSm}
                >
                  Edit
                </button>
                <button
                  onClick={() => setConfirmDel(l.Name)}
                  disabled={!!busy}
                  className={`${cls.btnSm} text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10`}
                >
                  {busy === `label-del-${l.Name}` ? "…" : "Delete"}
                </button>
              </span>
            </div>
          ))
        ) : (
          <div className="p-3 text-sm italic text-muted-foreground">no labels</div>
        )}
      </div>

      {confirmDel && (
        <ConfirmDialog
          title="Delete label?"
          body={
            <>
              Delete label <span className="text-foreground">{confirmDel}</span> from the repo? It
              is removed from any issues/PRs that use it.
            </>
          }
          confirmLabel="Delete"
          busy={!!busy}
          onCancel={() => setConfirmDel(null)}
          onConfirm={() => doDelete(confirmDel)}
        />
      )}
      {editing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-96 space-y-3 rounded-lg border border-border bg-card p-4 shadow-lg">
            <div className="text-sm font-semibold">Edit label</div>
            <input
              value={editing.newName}
              onChange={(e) => setEditing({ ...editing, newName: e.target.value })}
              placeholder="name"
              className={`${cls.input} w-full`}
            />
            <div className="flex items-center gap-2">
              <input
                type="color"
                value={`#${normHex(editing.Color)}`}
                onChange={(e) => setEditing({ ...editing, Color: normHex(e.target.value) })}
                className="h-8 w-10 shrink-0 cursor-pointer rounded border border-border bg-transparent p-0"
                title="pick a color"
              />
              <button
                type="button"
                onClick={() => setEditing({ ...editing, Color: randomHex() })}
                className={cls.btnSm}
                title="random color"
              >
                🎲
              </button>
              <input
                value={editing.Color}
                onChange={(e) => setEditing({ ...editing, Color: normHex(e.target.value) })}
                placeholder="hex"
                className={`${cls.input} w-24`}
              />
            </div>
            <input
              value={editing.Description}
              onChange={(e) => setEditing({ ...editing, Description: e.target.value })}
              placeholder="description"
              className={`${cls.input} w-full`}
            />
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setEditing(null)}
                disabled={!!busy}
                className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40"
              >
                Cancel
              </button>
              <button
                onClick={doEdit}
                disabled={!!busy || !editing.newName.trim()}
                className={cls.btn}
              >
                {busy === "label-edit" ? "…" : "Save"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
