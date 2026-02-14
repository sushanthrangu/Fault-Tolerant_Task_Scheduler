import { useState, useEffect, useSyncExternalStore } from "react";
import type { Job } from "@/types/job";
import { getApiCalls, subscribeApiCalls, type ApiCall } from "@/api/client";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

interface ObservabilityPanelProps {
  jobs: Job[];
}

export function ObservabilityPanel({ jobs }: ObservabilityPanelProps) {
  const apiCalls = useSyncExternalStore(subscribeApiCalls, getApiCalls);

  const counts = {
    PENDING: jobs.filter((j) => j.status === "PENDING").length,
    RUNNING: jobs.filter((j) => j.status === "RUNNING").length,
    SUCCESS: jobs.filter((j) => j.status === "SUCCESS").length,
    FAILED: jobs.filter((j) => j.status === "FAILED").length,
  };

  // Simple chart data: jobs by creation minute
  const chartData = (() => {
    const map = new Map<string, number>();
    jobs.forEach((j) => {
      const d = new Date(j.created_at);
      const key = `${d.getHours()}:${String(d.getMinutes()).padStart(2, "0")}`;
      map.set(key, (map.get(key) || 0) + 1);
    });
    return Array.from(map.entries())
      .map(([time, count]) => ({ time, count }))
      .slice(-15);
  })();

  const statCard = (
    label: string,
    value: number,
    cls: string
  ) => (
    <div className={`rounded-lg border px-4 py-3 ${cls}`}>
      <p className="text-2xl font-bold font-mono">{value}</p>
      <p className="text-xs mt-0.5 opacity-70">{label}</p>
    </div>
  );

  return (
    <div className="bg-card border border-border rounded-lg p-5 card-glow space-y-5">
      <h2 className="text-sm font-semibold text-foreground uppercase tracking-wider">
        Observability
      </h2>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {statCard("Pending", counts.PENDING, "status-pending")}
        {statCard("Running", counts.RUNNING, "status-running")}
        {statCard("Success", counts.SUCCESS, "status-success")}
        {statCard("Failed", counts.FAILED, "status-failed")}
      </div>

      {chartData.length > 1 && (
        <div>
          <p className="text-xs text-muted-foreground mb-2">Jobs Over Time</p>
          <ResponsiveContainer width="100%" height={120}>
            <LineChart data={chartData}>
              <XAxis
                dataKey="time"
                tick={{ fontSize: 10, fill: "hsl(215 15% 50%)" }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis hide allowDecimals={false} />
              <Tooltip
                contentStyle={{
                  background: "hsl(220 18% 12%)",
                  border: "1px solid hsl(220 14% 18%)",
                  borderRadius: "6px",
                  fontSize: 12,
                }}
              />
              <Line
                type="monotone"
                dataKey="count"
                stroke="hsl(142 60% 45%)"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      <div>
        <p className="text-xs text-muted-foreground mb-2">
          Recent API Calls ({apiCalls.length})
        </p>
        <div className="space-y-1 max-h-48 overflow-y-auto">
          {apiCalls.length === 0 && (
            <p className="text-xs text-muted-foreground/60">No calls yet</p>
          )}
          {apiCalls
            .slice()
            .reverse()
            .map((c, i) => (
              <div
                key={i}
                className="flex items-center gap-3 text-xs font-mono py-1 px-2 rounded bg-muted/40"
              >
                <span className="text-muted-foreground w-10 shrink-0">
                  {c.method}
                </span>
                <span className="text-foreground truncate flex-1">
                  {c.path}
                </span>
                <span
                  className={`w-8 text-right ${
                    c.status && c.status < 400
                      ? "text-primary"
                      : "text-destructive"
                  }`}
                >
                  {c.status ?? "ERR"}
                </span>
                <span className="text-muted-foreground w-12 text-right">
                  {c.duration}ms
                </span>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}
