import { useState } from "react";
import { useGit, cls } from "../context";
import { RebaseAction } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";
import type { RebaseStep } from "../../bindings/github.com/davasorus/gitmate/internal/gitops";

// RebasePanel drives an interactive rebase: pick how many commits back, then
// reorder / set an action (pick/reword/squash/fixup/drop) per commit, supply
// messages where needed, and run it. Nothing happens until "Start rebase".
export function RebasePanel({ onClose }: { onClose: () => void }) {
  const { busy, service, run, flash, setBusy } = useGit();
  const [count, setCount] = useState(5);
  const [base, setBase] = useState("");
  const [steps, setSteps] = useState<RebaseStep[]>([]);
  const [loaded, setLoaded] = useState(false);

  const load = async () => {
    setBusy("rebase-todo");
    try {
      // base = HEAD~count; the engine lists base..HEAD as the editable plan
      const b = `HEAD~${count}`;
      const todo = (await service.InteractiveRebaseTodo(b)) ?? [];
      setBase(b);
      setSteps(todo);
      setLoaded(true);
    } catch (e) {
      flash("err", String(e));
    } finally {
      setBusy("");
    }
  };

  const move = (i: number, dir: -1 | 1) => {
    const j = i + dir;
    if (j < 0 || j >= steps.length) return;
    const next = steps.slice();
    [next[i], next[j]] = [next[j], next[i]];
    setSteps(next);
  };

  const setAction = (i: number, action: RebaseAction) => {
    const next = steps.slice();
    next[i] = { ...next[i], Action: action };
    setSteps(next);
  };

  const setMessage = (i: number, msg: string) => {
    const next = steps.slice();
    next[i] = { ...next[i], Message: msg };
    setSteps(next);
  };

  const start = () =>
    run(
      "rebase-run",
      async () => {
        await service.RunInteractiveRebase(base, steps);
        onClose();
        return "rebase complete";
      },
      "rebase complete",
    );

  const actions: RebaseAction[] = [
    RebaseAction.RebasePick,
    RebaseAction.RebaseReword,
    RebaseAction.RebaseSquash,
    RebaseAction.RebaseFixup,
    RebaseAction.RebaseDrop,
  ];

  return (
    <div className="space-y-2 rounded-lg border border-[var(--color-modified)] bg-[var(--color-modified)]/5 p-3">
      <div className="flex items-center gap-2">
        <span className="text-sm font-semibold">Interactive rebase</span>
        <button onClick={onClose} className={`${cls.btnSm} ml-auto`}>
          Cancel
        </button>
      </div>

      {!loaded ? (
        <div className="flex items-center gap-2 text-sm">
          <span>Last</span>
          <input
            type="number"
            min={1}
            max={50}
            value={count}
            onChange={(e) => setCount(Math.max(1, parseInt(e.target.value, 10) || 1))}
            className={`${cls.input} h-7 w-16 px-2 py-0`}
          />
          <span>commits</span>
          <button onClick={load} disabled={!!busy} className={cls.btn}>
            {busy === "rebase-todo" ? "…" : "Load"}
          </button>
        </div>
      ) : (
        <div className="space-y-2">
          <div className="text-xs text-muted-foreground">
            Reorder with ↑/↓, set an action, add a message for reword/squash. Applied top-to-bottom
            onto <span className="font-mono">{base}</span>.
          </div>
          {steps.map((s, i) => (
            <div
              key={s.SHA}
              className="flex items-center gap-2 rounded border border-border p-1.5 text-xs"
            >
              <span className="flex flex-col">
                <button
                  onClick={() => move(i, -1)}
                  disabled={i === 0}
                  className="disabled:opacity-30"
                >
                  ↑
                </button>
                <button
                  onClick={() => move(i, 1)}
                  disabled={i === steps.length - 1}
                  className="disabled:opacity-30"
                >
                  ↓
                </button>
              </span>
              <span className="font-mono text-[var(--color-modified)]">{s.SHA.slice(0, 7)}</span>
              <select
                value={s.Action}
                onChange={(e) => setAction(i, e.target.value as RebaseAction)}
                className={`${cls.input} h-6 px-1 py-0`}
              >
                {actions.map((a) => (
                  <option key={a} value={a}>
                    {a}
                  </option>
                ))}
              </select>
              <span
                className={`flex-1 truncate ${s.Action === RebaseAction.RebaseDrop ? "line-through opacity-50" : ""}`}
              >
                {s.Subject}
              </span>
              {(s.Action === RebaseAction.RebaseReword ||
                s.Action === RebaseAction.RebaseSquash) && (
                <input
                  value={s.Message ?? ""}
                  onChange={(e) => setMessage(i, e.target.value)}
                  placeholder={
                    s.Action === RebaseAction.RebaseReword ? "new message" : "combined message"
                  }
                  className={`${cls.input} h-6 w-40 px-2 py-0`}
                />
              )}
            </div>
          ))}
          <button onClick={start} disabled={!!busy} className={cls.btn}>
            {busy === "rebase-run" ? "…" : "Start rebase"}
          </button>
        </div>
      )}
    </div>
  );
}
