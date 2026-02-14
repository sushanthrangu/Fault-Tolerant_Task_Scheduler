import { useState } from "react";
import { createJob } from "@/api/client";
import type { Job } from "@/types/job";
import { useToast } from "@/hooks/use-toast";

interface CreateJobPanelProps {
  onJobCreated: (job: Job) => void;
}

export function CreateJobPanel({ onJobCreated }: CreateJobPanelProps) {
  const [type, setType] = useState("demo");
  const [payload, setPayload] = useState('{ "msg": "hello" }');
  const [maxAttempts, setMaxAttempts] = useState(3);
  const [idempotencyKey, setIdempotencyKey] = useState("");
  const [loading, setLoading] = useState(false);
  const [payloadError, setPayloadError] = useState("");
  const { toast } = useToast();

  const validatePayload = (val: string) => {
    try {
      JSON.parse(val);
      setPayloadError("");
      return true;
    } catch {
      setPayloadError("Invalid JSON");
      return false;
    }
  };

  const submit = async (autoKey = false) => {
    if (!validatePayload(payload)) return;

    const key = autoKey
      ? crypto.randomUUID()
      : idempotencyKey.trim() || undefined;

    if (autoKey && key) {
      setIdempotencyKey(key);
    }

    // Optimistic: add PENDING job immediately
    const optimisticJob: Job = {
      id: `pending-${Date.now()}`,
      type,
      payload: JSON.parse(payload),
      status: "PENDING",
      attempts: 0,
      max_attempts: maxAttempts,
      next_run_at: new Date().toISOString(),
      idempotency_key: key,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    onJobCreated(optimisticJob);

    setLoading(true);
    try {
      const job = await createJob(
        { type, payload: JSON.parse(payload), max_attempts: maxAttempts },
        key
      );
      // Replace optimistic with real
      onJobCreated(job);
      toast({
        title: "Job created",
        description: `ID: ${job.id.slice(0, 12)}…`,
      });
    } catch (err: unknown) {
      toast({
        title: "Failed to create job",
        description: err instanceof Error ? err.message : "Unknown error",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-card border border-border rounded-lg p-5 card-glow">
      <h2 className="text-sm font-semibold text-foreground mb-4 uppercase tracking-wider">
        Create Job
      </h2>
      <div className="grid gap-4 sm:grid-cols-2">
        <div>
          <label className="block text-xs text-muted-foreground mb-1">
            Job Type
          </label>
          <input
            className="w-full bg-muted border border-border rounded px-3 py-2 text-sm font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
            value={type}
            onChange={(e) => setType(e.target.value)}
          />
        </div>
        <div>
          <label className="block text-xs text-muted-foreground mb-1">
            Max Attempts
          </label>
          <input
            type="number"
            min={1}
            max={10}
            className="w-full bg-muted border border-border rounded px-3 py-2 text-sm font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
            value={maxAttempts}
            onChange={(e) =>
              setMaxAttempts(Math.max(1, Math.min(10, Number(e.target.value))))
            }
          />
        </div>
        <div className="sm:col-span-2">
          <label className="block text-xs text-muted-foreground mb-1">
            Payload (JSON)
          </label>
          <textarea
            rows={3}
            className={`w-full bg-muted border rounded px-3 py-2 text-sm font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-none ${
              payloadError ? "border-destructive" : "border-border"
            }`}
            value={payload}
            onChange={(e) => {
              setPayload(e.target.value);
              validatePayload(e.target.value);
            }}
          />
          {payloadError && (
            <p className="text-xs text-destructive mt-1">{payloadError}</p>
          )}
        </div>
        <div className="sm:col-span-2">
          <label className="block text-xs text-muted-foreground mb-1">
            Idempotency Key{" "}
            <span className="text-muted-foreground/60">(optional)</span>
          </label>
          <input
            className="w-full bg-muted border border-border rounded px-3 py-2 text-sm font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
            value={idempotencyKey}
            onChange={(e) => setIdempotencyKey(e.target.value)}
            placeholder="e.g. my-unique-key-123"
          />
        </div>
      </div>
      <div className="flex gap-3 mt-5">
        <button
          onClick={() => submit(false)}
          disabled={loading}
          className="px-4 py-2 bg-primary text-primary-foreground text-sm font-medium rounded hover:bg-primary/90 transition-colors disabled:opacity-50"
        >
          {loading ? "Creating…" : "Create Job"}
        </button>
        <button
          onClick={() => submit(true)}
          disabled={loading}
          className="px-4 py-2 bg-secondary text-secondary-foreground text-sm font-medium rounded hover:bg-secondary/80 transition-colors disabled:opacity-50 border border-border"
        >
          Create (Random Key)
        </button>
      </div>
    </div>
  );
}
