import { useState } from "react";
import { useGit, cls } from "../context";
import { ConfirmDialog } from "../components/ConfirmDialog";
import type { Tag } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";
import { Monitor, Cloud } from "lucide-react";

function LocationIcons({ t }: { t: Tag }) {
  const title =
    t.Local && t.Remote
      ? "local and origin"
      : t.Local
        ? "local only (not pushed)"
        : t.Remote
          ? "on origin only"
          : "";
  return (
    <span className="flex shrink-0 items-center gap-1" title={title}>
      <Monitor
        size={13}
        className={t.Local ? "text-[var(--color-modified)]" : "text-[var(--color-faint)]/30"}
      />
      <Cloud
        size={13}
        className={t.Remote ? "text-[var(--color-accent)]" : "text-[var(--color-faint)]/30"}
      />
    </span>
  );
}

export function Tags() {
  const { tags, busy, run, service } = useGit();
  const [name, setName] = useState("");
  const [msg, setMsg] = useState("");
  const [showNew, setShowNew] = useState(false);
  const [confirmDel, setConfirmDel] = useState<Tag | null>(null);

  const doCreate = () =>
    run(
      "tag-create",
      async () => {
        await service.CreateTag(name.trim(), msg.trim());
        setName("");
        setMsg("");
        setShowNew(false);
        return `created ${name.trim()}`;
      },
      "tag created",
    );
  const doPush = (n: string) =>
    run(
      `tag-push-${n}`,
      async () => {
        await service.PushTag(n);
        return `pushed ${n} — release workflow may trigger`;
      },
      `pushed ${n}`,
    );
  const doSync = () => run("tag-sync", () => service.FetchTags(), "synced tags from origin");
  const doDelete = (t: Tag) =>
    run(
      `tag-del-${t.Name}`,
      async () => {
        const where = await service.SmartDeleteTag(t.Name);
        setConfirmDel(null);
        return `deleted ${t.Name} (${where})`;
      },
      `deleted ${t.Name}`,
    );
  const delScope = (t: Tag) =>
    t.Local && t.Remote ? "from local and origin" : t.Local ? "locally" : "from origin";

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div className="flex items-center gap-2">
        <span className="text-[15px] font-semibold tracking-[-0.01em]">Tags</span>
        <button
          onClick={doSync}
          disabled={!!busy}
          title="fetch tags from origin, prune deleted ones"
          className="ml-auto rounded-lg border border-border px-2.5 py-1 text-[12px] font-medium text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] disabled:opacity-40"
        >
          {busy === "tag-sync" ? "…" : "Sync tags"}
        </button>
        <button
          onClick={() => setShowNew((v) => !v)}
          className="text-[11.5px] font-semibold text-[var(--color-accent)] hover:underline"
        >
          {showNew ? "Cancel" : "+ New tag"}
        </button>
      </div>

      {showNew && (
        <div className="space-y-2 rounded-lg border border-border p-3">
          <div className="text-[12px] text-[var(--color-muted-foreground)]">
            Tag the current commit (HEAD). Add a message for an annotated tag.
          </div>
          <div className="flex gap-2">
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="tag name (e.g. v1.2.0)"
              className={`${cls.input} flex-1`}
            />
            <button
              onClick={doCreate}
              disabled={!!busy || !name.trim()}
              className="rounded-lg bg-[var(--color-accent)] px-3.5 py-1.5 text-[12.5px] font-semibold text-white disabled:opacity-40"
            >
              {busy === "tag-create" ? "…" : "Create"}
            </button>
          </div>
          <textarea
            value={msg}
            onChange={(e) => setMsg(e.target.value)}
            placeholder="annotation message (optional) — first line is the subject"
            rows={2}
            className={`${cls.input} w-full resize-y`}
          />
        </div>
      )}

      <div className="space-y-1">
        {(tags ?? []).length ? (
          tags.map((t) => (
            <div
              key={t.Name}
              className="group flex items-center gap-2.5 rounded-lg border border-border px-3 py-2.5 hover:bg-[var(--color-card)]"
            >
              <span className="shrink-0 font-mono text-[12px] font-medium text-[var(--color-modified)]">
                {t.Name}
              </span>
              <LocationIcons t={t} />
              <span className="min-w-0 flex-1 truncate text-[12.5px] text-[var(--color-muted-foreground)]">
                {t.Subject}
              </span>
              <span className="flex shrink-0 gap-1 opacity-0 group-hover:opacity-100">
                {t.Local && !t.Remote && (
                  <button
                    onClick={() => doPush(t.Name)}
                    disabled={!!busy}
                    title="push tag to origin (triggers release)"
                    className="rounded-md px-2 py-1 text-[11.5px] text-[var(--color-accent)] hover:bg-[var(--color-muted)]"
                  >
                    {busy === `tag-push-${t.Name}` ? "…" : "Push"}
                  </button>
                )}
                <button
                  onClick={() => setConfirmDel(t)}
                  disabled={!!busy}
                  className="rounded-md px-2 py-1 text-[11.5px] text-[var(--color-removed)] hover:bg-[var(--color-removed)]/10"
                >
                  {busy === `tag-del-${t.Name}` ? "…" : "Delete"}
                </button>
              </span>
            </div>
          ))
        ) : (
          <div className="rounded-lg border border-border px-3 py-6 text-center text-[12.5px] italic text-[var(--color-faint)]">
            no tags
          </div>
        )}
      </div>

      {confirmDel && (
        <ConfirmDialog
          title="Delete tag?"
          body={
            <>
              Delete tag <span className="text-foreground">{confirmDel.Name}</span>{" "}
              {delScope(confirmDel)}.
            </>
          }
          confirmLabel="Delete"
          busy={!!busy}
          onCancel={() => setConfirmDel(null)}
          onConfirm={() => doDelete(confirmDel)}
        />
      )}
    </div>
  );
}
