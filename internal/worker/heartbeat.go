package worker

import (
	"context"
	"time"

	mysqlrepo "task-scheduler/internal/repo/mysql"
)

type HeartbeatManager struct {
	Repo     *mysqlrepo.JobRepo
	WorkerID string
	ExtendBy time.Duration
	Interval time.Duration
	Now      func() time.Time
}

func NewHeartbeatManager(repo *mysqlrepo.JobRepo, workerID string, extendBy, interval time.Duration) *HeartbeatManager {
	return &HeartbeatManager{
		Repo:     repo,
		WorkerID: workerID,
		ExtendBy: extendBy,
		Interval: interval,
		Now:      time.Now,
	}
}

// Run periodically extends lease for a running job (call per job if you track running jobs).
func (h *HeartbeatManager) Run(ctx context.Context, jobID string) error {
	t := time.NewTicker(h.Interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			_ = h.Repo.Heartbeat(ctx, jobID, h.WorkerID, h.ExtendBy, h.Now())
		}
	}
}
