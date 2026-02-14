import { useState, useEffect, useCallback } from "react";
import type { Job } from "@/types/job";
import { loadRecentJobs, saveRecentJobs } from "@/utils/storage";
import { Header } from "@/components/Header";
import { CreateJobPanel } from "@/components/CreateJobPanel";
import { RecentJobsTable } from "@/components/RecentJobsTable";
import { JobDetailModal } from "@/components/JobDetailModal";
import { ObservabilityPanel } from "@/components/ObservabilityPanel";

const Index = () => {
  const [jobs, setJobs] = useState<Job[]>(() => loadRecentJobs());
  const [viewingJob, setViewingJob] = useState<Job | null>(null);

  useEffect(() => {
    saveRecentJobs(jobs);
  }, [jobs]);

  const handleJobCreated = useCallback((job: Job) => {
    setJobs((prev) => {
      // Replace optimistic or existing job with same id or pending prefix
      const filtered = prev.filter(
        (j) =>
          j.id !== job.id &&
          !(j.id.startsWith("pending-") && job.idempotency_key && j.idempotency_key === job.idempotency_key)
      );
      return [job, ...filtered];
    });
  }, []);

  const handleJobUpdate = useCallback((job: Job) => {
    setJobs((prev) => prev.map((j) => (j.id === job.id ? job : j)));
    setViewingJob((v) => (v && v.id === job.id ? job : v));
  }, []);

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-7xl px-4 sm:px-6 py-6 space-y-6">
        <div className="grid lg:grid-cols-5 gap-6">
          <div className="lg:col-span-3">
            <CreateJobPanel onJobCreated={handleJobCreated} />
          </div>
          <div className="lg:col-span-2">
            <ObservabilityPanel jobs={jobs} />
          </div>
        </div>
        <RecentJobsTable
          jobs={jobs}
          onUpdate={handleJobUpdate}
          onView={setViewingJob}
        />
      </main>
      {viewingJob && (
        <JobDetailModal
          job={viewingJob}
          onClose={() => setViewingJob(null)}
          onUpdate={handleJobUpdate}
        />
      )}
    </div>
  );
};

export default Index;
