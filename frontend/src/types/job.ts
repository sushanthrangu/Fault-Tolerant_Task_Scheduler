export interface Job {
  id: string;
  type: string;
  payload: Record<string, unknown>;
  status: "PENDING" | "RUNNING" | "SUCCESS" | "FAILED";
  attempts: number;
  max_attempts: number;
  next_run_at: string;
  idempotency_key?: string;
  created_at: string;
  updated_at: string;
  locked_by?: string;
  locked_until?: string;
  started_at?: string;
  completed_at?: string;
  error_message?: string;
}

export interface CreateJobRequest {
  type: string;
  payload: Record<string, unknown>;
  max_attempts: number;
}
