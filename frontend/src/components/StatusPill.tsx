import type { Job } from "@/types/job";

export function StatusPill({ status }: { status: Job["status"] }) {
  const cls =
    status === "PENDING"
      ? "status-pending"
      : status === "RUNNING"
      ? "status-running"
      : status === "SUCCESS"
      ? "status-success"
      : "status-failed";

  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs font-mono font-medium ${cls}`}>
      {status}
    </span>
  );
}
