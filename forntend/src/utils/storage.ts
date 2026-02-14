import type { Job } from "@/types/job";

const STORAGE_KEY = "ftts_recent_jobs";

export function loadRecentJobs(): Job[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    return JSON.parse(raw) as Job[];
  } catch {
    return [];
  }
}

export function saveRecentJobs(jobs: Job[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(jobs.slice(0, 50)));
}
