import { useState, useEffect, useCallback } from "react";
import { checkHealth } from "@/api/client";

export function useHealthCheck(intervalMs = 5000) {
  const [healthy, setHealthy] = useState<boolean | null>(null);

  const check = useCallback(async () => {
    const ok = await checkHealth();
    setHealthy(ok);
  }, []);

  useEffect(() => {
    check();
    const id = setInterval(check, intervalMs);
    return () => clearInterval(id);
  }, [check, intervalMs]);

  return healthy;
}
