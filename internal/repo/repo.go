package repo

import (
	"context"
	"time"

	"task-scheduler/internal/domain"
)

type JobRepository interface {
	// API operations
	CreateJob(ctx context.Context, id, jobType string, payload []byte, maxAttempts int, idempotencyKey *string) (*domain.Job, error)
	GetJobByID(ctx context.Context, id string) (*domain.Job, error)
	GetJobByIdempotencyKey(ctx context.Context, key string) (*domain.Job, error)

	// Worker operations
	// ClaimJobs atomically "leases" jobs for this worker to execute.
	// It should return jobs already moved to RUNNING with locked_by/locked_until set.
	ClaimJobs(ctx context.Context, workerID string, limit int, lease time.Duration, now time.Time) ([]domain.Job, error)

	// Heartbeat extends the lease for long-running jobs (optional but production-grade).
	Heartbeat(ctx context.Context, jobID string, workerID string, extendBy time.Duration, now time.Time) error

	// State transitions
	MarkSuccess(ctx context.Context, jobID string, completedAt time.Time) error
	MarkFailure(ctx context.Context, jobID string, attempts int, nextRunAt *time.Time, errMsg string, terminal bool, completedAt *time.Time) error

	// Execution idempotency for side-effects (optional now, but we’ll use it soon)
	RecordStepOnce(ctx context.Context, jobID string, stepKey string, resultHash *string) (inserted bool, err error)
}
