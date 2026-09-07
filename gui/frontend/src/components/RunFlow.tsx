import type { JobNode, Job } from "../../bindings/github.com/davasorus/gitmate/internal/ghapi";

// RunFlow draws the workflow's job dependency graph (from needs:) as columns by
// depth. Matrix legs (nodes named "base (leg)") are GROUPED under a single
// parent node, with each leg shown as a row inside and colored by its own live
// status — so "go" matrixed 3 ways reads as one job, not three.
export function RunFlow({ graph, jobs }: { graph: JobNode[]; jobs: Job[] }) {
  if (!graph || graph.length === 0) {
    return (
      <div className="text-xs italic text-[var(--color-faint)]">
        no job graph (workflow YAML unavailable or no jobs)
      </div>
    );
  }

  const statusOf = (name: string) => {
    const j =
      (jobs ?? []).find((x) => x.Name === name) ??
      (jobs ?? []).find((x) => x.Name.startsWith(name));
    if (!j) return { status: "", conclusion: "" };
    return { status: j.Status, conclusion: j.Conclusion };
  };
  const color = (name: string) => {
    const { status, conclusion } = statusOf(name);
    const s = conclusion || status;
    if (s === "success") return "var(--color-added)";
    if (s === "failure" || s === "cancelled") return "var(--color-removed)";
    if (s === "in_progress" || s === "queued") return "var(--color-behind)";
    return "var(--color-faint, #6a717c)";
  };

  // --- group matrix legs under a base job ---------------------------------
  // Engine expands matrix into nodes named "base (leg1, leg2)". Cluster them by
  // base, so the graph draws one parent box per base with N leg-rows inside.
  const splitLeg = (name: string): { base: string; leg: string | null } => {
    const m = name.match(/^(.*) \((.*)\)$/);
    return m ? { base: m[1], leg: m[2] } : { base: name, leg: null };
  };

  type Group = { base: string; legs: JobNode[]; needsBases: string[] };
  const groupMap = new Map<string, Group>();
  const baseOrder: string[] = [];
  for (const n of graph) {
    const { base } = splitLeg(n.Name);
    if (!groupMap.has(base)) {
      groupMap.set(base, { base, legs: [], needsBases: [] });
      baseOrder.push(base);
    }
    const g = groupMap.get(base)!;
    g.legs.push(n);
    for (const dep of n.Needs ?? []) {
      const db = splitLeg(dep).base;
      if (!g.needsBases.includes(db)) g.needsBases.push(db);
    }
  }
  const groups = baseOrder.map((b) => groupMap.get(b)!);

  // depth by base (longest path over base-level needs)
  const gByBase = new Map(groups.map((g) => [g.base, g]));
  const depthCache = new Map<string, number>();
  const depth = (base: string, seen = new Set<string>()): number => {
    if (depthCache.has(base)) return depthCache.get(base)!;
    if (seen.has(base)) return 0;
    seen.add(base);
    const g = gByBase.get(base);
    const needs = g?.needsBases.filter((b) => b !== base) ?? [];
    const d = needs.length === 0 ? 0 : 1 + Math.max(...needs.map((b) => depth(b, seen)));
    depthCache.set(base, d);
    return d;
  };

  const columns: Group[][] = [];
  for (const g of groups) (columns[depth(g.base)] ||= []).push(g);

  // geometry — parent box height grows with leg count
  const colW = 240;
  const boxW = 200;
  const headH = 26; // parent header
  const legH = 22; // per leg row
  const gapY = 22;
  const padX = 16;
  const padY = 16;
  const boxHeight = (g: Group) => headH + (g.legs.length > 1 ? g.legs.length * legH + 6 : 0);

  const pos = new Map<string, { x: number; y: number; h: number }>();
  columns.forEach((col, ci) => {
    let y = padY;
    col.forEach((g) => {
      const h = boxHeight(g);
      pos.set(g.base, { x: padX + ci * colW, y, h });
      y += h + gapY;
    });
  });
  const width = padX * 2 + columns.length * colW;
  const height =
    padY * 2 +
    Math.max(...columns.map((col) => col.reduce((s, g) => s + boxHeight(g) + gapY, 0)), 60);

  const edgeColor = "var(--color-border-strong, #555)";

  return (
    <div className="overflow-auto">
      <svg width={width} height={height} className="min-w-full">
        <defs>
          <marker id="arrow" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">
            <path d="M0,0 L8,4 L0,8 Z" fill={edgeColor} />
          </marker>
        </defs>

        {/* base-level dependency edges (parent → parent) */}
        {groups.flatMap((g) => {
          const to = pos.get(g.base);
          if (!to) return [];
          return g.needsBases
            .filter((b) => b !== g.base && pos.has(b))
            .map((b, i) => {
              const from = pos.get(b)!;
              const x1 = from.x + boxW;
              const y1 = from.y + Math.min(from.h, 40) / 2;
              const x2 = to.x;
              const y2 = to.y + Math.min(to.h, 40) / 2;
              return (
                <path
                  key={`${g.base}-${b}-${i}`}
                  d={`M${x1},${y1} C${x1 + 40},${y1} ${x2 - 40},${y2} ${x2},${y2}`}
                  fill="none"
                  stroke={edgeColor}
                  strokeWidth={1.5}
                  markerEnd="url(#arrow)"
                />
              );
            });
        })}

        {/* group boxes */}
        {groups.map((g) => {
          const p = pos.get(g.base)!;
          const isMatrix = g.legs.length > 1;
          // parent color: worst-of the legs (fail > running > pending > pass)
          const legColors = g.legs.map((l) => color(l.Name));
          const parentColor = isMatrix
            ? legColors.includes("var(--color-removed)")
              ? "var(--color-removed)"
              : legColors.includes("var(--color-behind)")
                ? "var(--color-behind)"
                : legColors.every((c) => c === "var(--color-added)")
                  ? "var(--color-added)"
                  : "var(--color-border-strong, #555)"
            : color(g.legs[0].Name);
          return (
            <g key={g.base}>
              <rect
                x={p.x}
                y={p.y}
                width={boxW}
                height={p.h}
                rx={9}
                fill="var(--color-card)"
                stroke={parentColor}
                strokeWidth={1.5}
              />
              {/* header */}
              <circle cx={p.x + 13} cy={p.y + headH / 2 + 1} r={4} fill={parentColor} />
              <text
                x={p.x + 24}
                y={p.y + headH / 2 + 5}
                fontSize={12}
                fontWeight={510}
                fill="var(--color-foreground)"
                fontFamily="var(--font-sans)"
              >
                {g.base.length > 20 ? g.base.slice(0, 19) + "…" : g.base}
              </text>
              {isMatrix && (
                <text
                  x={p.x + boxW - 10}
                  y={p.y + headH / 2 + 5}
                  fontSize={9}
                  textAnchor="end"
                  fill="var(--color-faint)"
                  fontFamily="var(--font-mono)"
                >
                  ×{g.legs.length}
                </text>
              )}
              {/* leg rows (only when matrixed) */}
              {isMatrix &&
                g.legs.map((leg, li) => {
                  const { leg: legLabel } = splitLeg(leg.Name);
                  const lc = color(leg.Name);
                  const ry = p.y + headH + 4 + li * legH;
                  return (
                    <g key={leg.Name}>
                      <rect
                        x={p.x + 8}
                        y={ry}
                        width={boxW - 16}
                        height={legH - 3}
                        rx={5}
                        fill="transparent"
                        stroke={lc}
                        strokeWidth={1}
                        opacity={0.85}
                      />
                      <text
                        x={p.x + 16}
                        y={ry + (legH - 3) / 2 + 4}
                        fontSize={10}
                        fill="var(--color-muted-foreground)"
                        fontFamily="var(--font-mono)"
                      >
                        {(legLabel ?? "").length > 22 ? legLabel!.slice(0, 21) + "…" : legLabel}
                      </text>
                    </g>
                  );
                })}
            </g>
          );
        })}
      </svg>
    </div>
  );
}
