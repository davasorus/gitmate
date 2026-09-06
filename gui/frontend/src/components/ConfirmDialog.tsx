import { ReactNode } from "react";

// Reusable confirm modal. `danger` styles the primary button as destructive.
// `extraAction` optionally renders a second confirming button (e.g. force delete).
export function ConfirmDialog({
  title,
  body,
  confirmLabel,
  onConfirm,
  onCancel,
  busy,
  danger = true,
  extra,
}: {
  title: string;
  body: ReactNode;
  confirmLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
  busy: boolean;
  danger?: boolean;
  extra?: { label: string; onClick: () => void };
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-96 rounded-lg border border-border bg-card p-4 shadow-lg">
        <div
          className={`mb-2 text-sm font-semibold ${danger ? "text-[var(--color-removed)]" : ""}`}
        >
          {title}
        </div>
        <div className="mb-4 break-all text-xs text-muted-foreground">{body}</div>
        <div className="flex justify-end gap-2">
          <button
            onClick={onCancel}
            disabled={busy}
            className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40"
          >
            Cancel
          </button>
          {extra && (
            <button
              onClick={extra.onClick}
              disabled={busy}
              className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-muted disabled:opacity-40"
            >
              {extra.label}
            </button>
          )}
          <button
            onClick={onConfirm}
            disabled={busy}
            className={`rounded-md px-3 py-1.5 text-sm font-semibold disabled:opacity-40 ${danger ? "bg-[var(--color-removed)] text-white" : "bg-primary text-background"}`}
          >
            {busy ? "…" : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
