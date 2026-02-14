import type { Job, CreateJobRequest } from "@/types/job";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8086";

export function getApiBase() {
  return API_BASE;
}

export interface ApiCall {
  method: string;
  path: string;
  status: number | null;
  duration: number;
  timestamp: number;
  error?: string;
}

let apiCallLog: ApiCall[] = [];
let apiCallSnapshot: ApiCall[] = [];
let listeners: (() => void)[] = [];

export function getApiCalls() {
  return apiCallSnapshot;
}

export function subscribeApiCalls(fn: () => void) {
  listeners.push(fn);
  return () => {
    listeners = listeners.filter((l) => l !== fn);
  };
}

function notifyListeners() {
  listeners.forEach((fn) => fn());
}

function recordCall(call: ApiCall) {
  apiCallLog = [...apiCallLog.slice(-19), call];
  apiCallSnapshot = apiCallLog.slice(-10);
  notifyListeners();
}

async function apiFetch<T>(
  path: string,
  options?: RequestInit & { rawText?: boolean }
): Promise<T> {
  const url = `${API_BASE}${path}`;
  const start = performance.now();
  let status: number | null = null;

  try {
    const res = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
      },
    });
    status = res.status;
    const duration = Math.round(performance.now() - start);
    recordCall({
      method: options?.method || "GET",
      path,
      status,
      duration,
      timestamp: Date.now(),
    });

    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || `HTTP ${res.status}`);
    }

    if (options?.rawText) {
      return (await res.text()) as unknown as T;
    }
    return await res.json();
  } catch (err: unknown) {
    const duration = Math.round(performance.now() - start);
    const message =
      err instanceof TypeError
        ? "Network error – possible CORS block or backend unreachable. Ensure the backend is running and CORS is configured, or use a browser extension to allow cross-origin requests."
        : err instanceof Error
        ? err.message
        : "Unknown error";

    if (status === null) {
      recordCall({
        method: options?.method || "GET",
        path,
        status: null,
        duration,
        timestamp: Date.now(),
        error: message,
      });
    }
    throw new Error(message);
  }
}

export async function checkHealth(): Promise<boolean> {
  try {
    const text = await apiFetch<string>("/healthz", { rawText: true } as any);
    return text.trim() === "ok";
  } catch {
    return false;
  }
}

export async function createJob(
  req: CreateJobRequest,
  idempotencyKey?: string
): Promise<Job> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (idempotencyKey) {
    headers["Idempotency-Key"] = idempotencyKey;
  }
  return apiFetch<Job>("/jobs", {
    method: "POST",
    headers,
    body: JSON.stringify(req),
  });
}

export async function getJob(id: string): Promise<Job> {
  return apiFetch<Job>(`/jobs/${id}`);
}
