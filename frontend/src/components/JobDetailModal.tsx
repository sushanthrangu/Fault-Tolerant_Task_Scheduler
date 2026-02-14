import { useState, useEffect, useRef } from "react";
import type { Job } from "@/types/job";
import { getJob } from "@/api/client";
import { StatusPill } from "@/components/StatusPill";
import { X } from "lucide-react";

interface JobDetailModalProps {
  job: Job;
  onClose: () => void;
  onUpdate: (job: Job) => void;
}

export function JobDetailModal({ job, onClose, onUpdate }: JobDetailModalProps) {
  const [current, setCurrent] = useState(job);
  const [polling, setPolling] = useState(false);
  const intervalRef = useRef<number | null>(null);

  useEffect(() => {
    // Fetch latest on open
    if (!job.id.startsWith("pending-")) {
      getJob(job.id).then((j) => {
        setCurrent(j);
        onUpdate(j);
      }).catch(() => {});
    }
  }, [job.id]);

  useEffect(() => {
    if (polling && !job.id.startsWith("pending-")) {
      intervalRef.current = window.setInterval(async () => {
        try {
          const j = await getJob(job.id);
          setCurrent(j);
          onUpdate(j);
          if (j.status === "SUCCESS" || j.status === "FAILED") {
            setPolling(false);
          }
        } catch {}
      }, 1000);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [polling, job.id]);

  const fmt = (iso?: string) =>
    iso ? new Date(iso).toLocaleString() : "—";

  const field = (label: string, value: React.ReactNode) => (
    <div className="flex justify-between py-1.5 border-b border-border/30">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-xs text-foreground font-mono">{value}</span>
    </div>
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm" onClick={onClose}>
      <div
        className="bg-card border border-border rounded-lg w-full max-w-lg mx-4 max-h-[85vh] overflow-y-auto card-glow"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-5 py-3 border-b border-border">
          <h2 className="text-sm font-semibold text-foreground uppercase tracking-wider">
            Job Detail
          </h2>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors">
            <X size={18} />
          </button>
        </div>
        <div className="p-5 space-y-3">
          {field("Status", <StatusPill status={current.status} />)}
          {field("ID", current.id)}
          {field("Type", current.type)}
          {field("Attempts", `${current.attempts} / ${current.max_attempts}`)}
          {field("Idempotency Key", current.idempotency_key || "—")}
          {field("Next Run", fmt(current.next_run_at))}
          {field("Created", fmt(current.created_at))}
          {field("Updated", fmt(current.updated_at))}
          {current.locked_by && field("Locked By", current.locked_by)}
          {current.locked_until && field("Locked Until", fmt(current.locked_until))}
          {current.started_at && field("Started At", fmt(current.started_at))}
          {current.completed_at && field("Completed At", fmt(current.completed_at))}
          {current.error_message && (
            <div className="mt-2 p-3 bg-destructive/10 border border-destructive/20 rounded text-xs text-destructive font-mono">
              {current.error_message}
            </div>
          )}

          <div className="flex items-center justify-between pt-3">
            <span className="text-xs text-muted-foreground">Poll Status (1s)</span>
            <button
              onClick={() => setPolling((p) => !p)}
              disabled={job.id.startsWith("pending-")}
              className={`relative w-10 h-5 rounded-full transition-colors ${
                polling ? "bg-primary" : "bg-muted"
              }`}
            >
              <span
                className={`absolute top-0.5 w-4 h-4 rounded-full transition-transform ${
                  polling
                    ? "translate-x-5 bg-primary-foreground"
                    : "translate-x-0.5 bg-muted-foreground"
                }`}
              />
            </button>
          </div>

          <div className="pt-3">
            <p className="text-xs text-muted-foreground mb-2">Raw JSON</p>
            <pre className="bg-muted p-3 rounded text-xs font-mono text-foreground overflow-x-auto max-h-48">
              {JSON.stringify(current, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}
