import { getApiBase } from "@/api/client";
import { useHealthCheck } from "@/hooks/use-health-check";

export function Header() {
  const healthy = useHealthCheck();

  return (
    <header className="border-b border-border bg-card px-6 py-4">
      <div className="mx-auto max-w-7xl flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground">
            Fault-Tolerant Task Scheduler
          </h1>
          <p className="text-sm text-muted-foreground font-mono">
            Go + MySQL + Docker Compose
          </p>
        </div>
        <div className="flex items-center gap-4">
          <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded font-mono">
            {getApiBase()}
          </code>
          <div
            className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium ${
              healthy === null
                ? "bg-muted text-muted-foreground"
                : healthy
                ? "status-success glow-green"
                : "status-failed glow-red"
            }`}
          >
            <span
              className={`inline-block w-2 h-2 rounded-full ${
                healthy === null
                  ? "bg-muted-foreground"
                  : healthy
                  ? "bg-primary animate-pulse_dot"
                  : "bg-destructive animate-pulse_dot"
              }`}
            />
            {healthy === null ? "Checking…" : healthy ? "Healthy" : "Down"}
          </div>
        </div>
      </div>
    </header>
  );
}
