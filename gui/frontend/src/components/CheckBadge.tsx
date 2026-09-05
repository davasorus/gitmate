import type { CheckRun } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

export function CheckBadge({ run }: { run: CheckRun }) {
  const s = run.Conclusion || run.Status;
  const color =
    s === "success" ? "text-[var(--color-check-pass)]"
    : s === "failure" || s === "cancelled" || s === "timed_out" ? "text-[var(--color-check-fail)]"
    : "text-[var(--color-check-pending)]";
  return (
    <div className="text-xs">
      <span className={color}>{s}</span>
      <span className="ml-2 text-muted-foreground">{run.Name}</span>
    </div>
  );
}
