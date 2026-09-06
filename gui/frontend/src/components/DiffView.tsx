import type { FileDiff } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

export function DiffView({ files }: { files: FileDiff[] }) {
  if (!files || files.length === 0) {
    return <div className="p-3 text-xs italic text-muted-foreground">no diff</div>;
  }
  return (
    <div className="space-y-3">
      {files.map((f, fi) => (
        <div key={fi} className="overflow-hidden rounded-md border border-border">
          <div className="border-b border-border bg-muted px-3 py-1.5 text-xs text-[var(--color-ahead)]">
            {f.NewPath || f.OldPath}
          </div>
          {f.Binary ? (
            <div className="px-3 py-2 text-xs italic text-muted-foreground">binary file</div>
          ) : (
            (f.Hunks ?? []).map((h, hi) => (
              <div key={hi}>
                <div className="bg-muted/50 px-3 py-1 text-xs text-muted-foreground">
                  {h.Header}
                </div>
                <div className="overflow-x-auto">
                  {(h.Lines ?? []).map((ln, li) => {
                    const add = ln.Kind === "add";
                    const rem = ln.Kind === "remove";
                    const bg = add
                      ? "bg-[var(--color-added)]/10"
                      : rem
                        ? "bg-[var(--color-removed)]/10"
                        : "";
                    const fg = add
                      ? "text-[var(--color-added)]"
                      : rem
                        ? "text-[var(--color-removed)]"
                        : "text-foreground";
                    const marker = add ? "+" : rem ? "-" : " ";
                    return (
                      <div key={li} className={`flex ${bg} font-mono text-xs leading-5`}>
                        <span className="w-10 shrink-0 select-none px-1 text-right text-muted-foreground/60">
                          {ln.OldNum || ""}
                        </span>
                        <span className="w-10 shrink-0 select-none px-1 text-right text-muted-foreground/60">
                          {ln.NewNum || ""}
                        </span>
                        <span className={`w-4 shrink-0 select-none text-center ${fg}`}>
                          {marker}
                        </span>
                        <span className={`whitespace-pre ${fg}`}>{ln.Content}</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))
          )}
        </div>
      ))}
    </div>
  );
}
