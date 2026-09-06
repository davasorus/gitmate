import type { JobNode, Job } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

// RunFlow draws the workflow's job dependency graph (from needs:) as columns by
// depth, with arrows for dependencies, each box colored by the live job status.
export function RunFlow({ graph, jobs }: { graph: JobNode[]; jobs: Job[] }) {
  if (!graph || graph.length === 0) {
    return (
      <div className="text-xs italic text-muted-foreground">
        no job graph (workflow YAML unavailable or no jobs)
      </div>
    );
  }

  // live status per job name (from the run's jobs)
  const statusOf = (name: string) => {
    const j = (jobs ?? []).find((x) => x.Name === name || x.Name.startsWith(name));
    if (!j) return { status: "", conclusion: "" };
    return { status: j.Status, conclusion: j.Conclusion };
  };
  const color = (name: string) => {
    const { status, conclusion } = statusOf(name);
    const s = conclusion || status;
    if (s === "success") return "var(--color-added)";
    if (s === "failure" || s === "cancelled") return "var(--color-removed)";
    if (s === "in_progress" || s === "queued") return "var(--color-behind)";
    return "var(--color-muted-foreground, #888)";
  };

  // compute depth (longest path from a root) for column placement
  const byName = new Map(graph.map((n) => [n.Name, n]));
  const depthCache = new Map<string, number>();
  const depth = (name: string, seen = new Set<string>()): number => {
    if (depthCache.has(name)) return depthCache.get(name)!;
    if (seen.has(name)) return 0; // cycle guard
    seen.add(name);
    const node = byName.get(name);
    const needs = node?.Needs ?? [];
    const d = needs.length === 0 ? 0 : 1 + Math.max(...needs.map((n) => depth(n, seen)));
    depthCache.set(name, d);
    return d;
  };

  // group jobs into columns by depth
  const columns: JobNode[][] = [];
  for (const n of graph) {
    const d = depth(n.Name);
    (columns[d] ||= []).push(n);
  }

  // layout geometry
  const colW = 180,
    rowH = 56,
    boxW = 150,
    boxH = 34,
    padX = 20,
    padY = 16;
  const pos = new Map<string, { x: number; y: number }>();
  columns.forEach((col, ci) => {
    col.forEach((n, ri) => {
      pos.set(n.Name, { x: padX + ci * colW, y: padY + ri * rowH });
    });
  });
  const maxRows = Math.max(...columns.map((c) => c.length), 1);
  const width = padX * 2 + columns.length * colW;
  const height = padY * 2 + maxRows * rowH;

  return (
    <div className="overflow-auto">
      <svg width={width} height={height} className="min-w-full">
        {/* dependency arrows */}
        {graph.flatMap((n) =>
          (n.Needs ?? []).map((dep, di) => {
            const from = pos.get(dep),
              to = pos.get(n.Name);
            if (!from || !to) return null;
            const x1 = from.x + boxW,
              y1 = from.y + boxH / 2;
            const x2 = to.x,
              y2 = to.y + boxH / 2;
            return (
              <line
                key={`${n.Name}-${dep}-${di}`}
                x1={x1}
                y1={y1}
                x2={x2}
                y2={y2}
                stroke="var(--color-border, #444)"
                strokeWidth={1.5}
                markerEnd="url(#arrow)"
              />
            );
          }),
        )}
        <defs>
          <marker id="arrow" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">
            <path d="M0,0 L8,4 L0,8 Z" fill="var(--color-border, #444)" />
          </marker>
        </defs>
        {/* job boxes */}
        {graph.map((n) => {
          const p = pos.get(n.Name)!;
          const c = color(n.Name);
          return (
            <g key={n.Name}>
              <rect
                x={p.x}
                y={p.y}
                width={boxW}
                height={boxH}
                rx={6}
                fill="var(--color-card, #1a1a1a)"
                stroke={c}
                strokeWidth={2}
              />
              <circle cx={p.x + 12} cy={p.y + boxH / 2} r={4} fill={c} />
              <text
                x={p.x + 24}
                y={p.y + boxH / 2 + 4}
                fontSize={11}
                fill="var(--color-foreground, #ddd)"
              >
                {n.Name.length > 16 ? n.Name.slice(0, 15) + "…" : n.Name}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}
