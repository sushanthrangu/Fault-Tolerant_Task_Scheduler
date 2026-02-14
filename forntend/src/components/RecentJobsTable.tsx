import { useState } from "react";
import type { Job } from "@/types/job";
import { getJob } from "@/api/client";
import { StatusPill } from "@/components/StatusPill";
import { useToast } from "@/hooks/use-toast";
import { Copy, Eye, RefreshCw } from "lucide-react";

interface RecentJobsTableProps {
  jobs: Job[];
  onUpdate: (job: Job) => void;
  onView: (job: Job) => void;
}

export function RecentJobsTable({ jobs, onUpdate, onView }: RecentJobsTableProps) {
  const [refreshingId, setRefreshingId] = useState<string | null>(null);
  const { toast } = useToast();

  const refresh = async (id: string) => {
    if (id.startsWith("pending-")) return;
    setRefreshingId(id);
    try {
      const job = await getJob(id);
      onUpdate(job);
    } catch (err: unknown) {
      toast({
        title: "Failed to refresh",
        description: err instanceof Error ? err.message : "Unknown error",
        variant: "destructive",
      });
    } finally {
      setRefreshingId(null);
    }
  };

  const copyId = (id: string) => {
    navigator.clipboard.writeText(id);
    toast({ title: "Copied", description: id.slice(0, 16) + "…" });
  };

  const fmt = (iso: string) => {
    try {
      return new Date(iso).toLocaleString(undefined, {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });
    } catch {
      return iso;
    }
  };

  if (jobs.length === 0) {
    return (
      <div className="bg-card border border-border rounded-lg p-8 text-center text-muted-foreground text-sm">
        No jobs yet. Create one above.
      </div>
    );
  }

  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden card-glow">
      <div className="px-5 py-3 border-b border-border">
        <h2 className="text-sm font-semibold text-foreground uppercase tracking-wider">
          Recent Jobs
        </h2>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-xs text-muted-foreground uppercase tracking-wider">
              <th className="px-4 py-2 text-left">Job ID</th>
              <th className="px-4 py-2 text-left">Type</th>
              <th className="px-4 py-2 text-left">Status</th>
              <th className="px-4 py-2 text-left">Attempts</th>
              <th className="px-4 py-2 text-left hidden md:table-cell">Next Run</th>
              <th className="px-4 py-2 text-left hidden lg:table-cell">Created</th>
              <th className="px-4 py-2 text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {jobs.map((job) => (
              <tr
                key={job.id}
                className="border-b border-border/50 hover:bg-muted/30 transition-colors"
              >
                <td className="px-4 py-2.5 font-mono text-xs">
                  <span className="flex items-center gap-1">
                    {job.id.slice(0, 10)}…
                    <button
                      onClick={() => copyId(job.id)}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                    >
                      <Copy size={12} />
                    </button>
                  </span>
                </td>
                <td className="px-4 py-2.5 font-mono text-xs">{job.type}</td>
                <td className="px-4 py-2.5">
                  <StatusPill status={job.status} />
                </td>
                <td className="px-4 py-2.5 font-mono text-xs">
                  {job.attempts}/{job.max_attempts}
                </td>
                <td className="px-4 py-2.5 text-xs text-muted-foreground hidden md:table-cell">
                  {fmt(job.next_run_at)}
                </td>
                <td className="px-4 py-2.5 text-xs text-muted-foreground hidden lg:table-cell">
                  {fmt(job.created_at)}
                </td>
                <td className="px-4 py-2.5 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <button
                      onClick={() => onView(job)}
                      className="p-1.5 rounded text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                      title="View details"
                    >
                      <Eye size={14} />
                    </button>
                    <button
                      onClick={() => refresh(job.id)}
                      disabled={refreshingId === job.id || job.id.startsWith("pending-")}
                      className="p-1.5 rounded text-muted-foreground hover:text-foreground hover:bg-muted transition-colors disabled:opacity-30"
                      title="Refresh"
                    >
                      <RefreshCw
                        size={14}
                        className={refreshingId === job.id ? "animate-spin" : ""}
                      />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
