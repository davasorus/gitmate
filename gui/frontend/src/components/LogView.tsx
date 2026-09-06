// LogView renders GitHub Actions log text with coloring so failures stand out.
// Primary signal = GitHub's universal workflow-command markers (##[error] /
// ##[warning] / ##[notice]) — emitted by Actions itself regardless of language.
// Secondary = a general failure-word heuristic for tools that print raw errors.
// Group markers are dimmed. Marker prefixes are stripped for clean text.

type LineKind = "error" | "warning" | "notice" | "group" | "normal";

// GitHub prefixes every log line with an ISO timestamp: "2026-...Z <text>".
// Strip it for display but keep it out of matching.
function stripTimestamp(line: string): string {
  const m = line.match(/^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?(.*)$/);
  return m ? m[1] : line;
}

const failureWord = /\b(error|errors|failed|failure|fatal|panic|exception)\b/i;
const exitCode = /\bexit code\s+[1-9]/i;

function classify(text: string): { kind: LineKind; clean: string } {
  // GitHub workflow-command markers (universal, authoritative)
  const marker = text.match(/##\[(error|warning|notice|group|endgroup)\]\s?(.*)$/i);
  if (marker) {
    const tag = marker[1].toLowerCase();
    if (tag === "error") return { kind: "error", clean: marker[2] };
    if (tag === "warning") return { kind: "warning", clean: marker[2] };
    if (tag === "notice") return { kind: "notice", clean: marker[2] };
    if (tag === "group") return { kind: "group", clean: marker[2] };
    if (tag === "endgroup") return { kind: "group", clean: "" };
  }
  // secondary heuristic — general, not language-specific
  if (failureWord.test(text) || exitCode.test(text)) return { kind: "error", clean: text };
  return { kind: "normal", clean: text };
}

const kindClass: Record<LineKind, string> = {
  error: "text-[var(--color-removed)]",
  warning: "text-[var(--color-behind)]",
  notice: "text-[var(--color-ahead)]",
  group: "text-muted-foreground/50",
  normal: "text-muted-foreground",
};

export function LogView({ text }: { text: string }) {
  const lines = (text ?? "").split("\n");
  return (
    <pre className="whitespace-pre-wrap">
      {lines.map((raw, i) => {
        const stripped = stripTimestamp(raw);
        const { kind, clean } = classify(stripped);
        if (kind === "group" && clean === "") return null; // hide bare endgroup
        return (
          <div key={i} className={kindClass[kind]}>
            {clean || "\u00A0"}
          </div>
        );
      })}
    </pre>
  );
}
