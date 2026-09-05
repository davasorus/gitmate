export function StatusBadge({ staged, unstaged }: { staged: string; unstaged: string }) {
  const label = [staged, unstaged].filter(Boolean).join("/") || "•";
  const color =
    staged === "conflict" || unstaged === "conflict" ? "text-[var(--color-conflict)]"
    : staged === "added" ? "text-[var(--color-added)]"
    : staged === "deleted" || unstaged === "deleted" ? "text-[var(--color-removed)]"
    : "text-[var(--color-modified)]";
  return <span className={`w-20 shrink-0 text-xs ${color}`}>{label}</span>;
}
