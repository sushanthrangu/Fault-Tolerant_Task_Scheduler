package worker

import (
	"context"
	"log"
	"math/rand"
	"time"

	mysqlrepo "task-scheduler/internal/repo/mysql"
)

// Runner executes jobs and applies retry/backoff + exactly-once success guard.
type Runner struct {
	Repo      *mysqlrepo.JobRepo
	Backoff   BackoffConfig
	FailRate  float64 // demo-only (failure injection)
	Logger    *log.Logger
	StepKeyOK string // step key used for success marker
}

func NewRunner(repo *mysqlrepo.JobRepo, backoff BackoffConfig, failRate float64, logger *log.Logger) *Runner {
	if logger == nil {
		logger = log.Default()
	}
	return &Runner{
		Repo:      repo,
		Backoff:   backoff,
		FailRate:  failRate,
		Logger:    logger,
		StepKeyOK: "execute_success",
	}
}

// Process implements Pool Handler interface.
func (r *Runner) Process(ctx context.Context, job Job) {
	start := time.Now()

	// Simulated work (replace later with handler registry by job type)
	failed := rand.Float64() < r.FailRate

	if !failed {
		// Exactly-once marker for completed side-effect.
		inserted, err := r.Repo.RecordStepOnce(ctx, job.ID, r.StepKeyOK, nil)
		if err != nil {
			r.Logger.Printf("job %s RecordStepOnce error: %v", job.ID, err)
			r.scheduleRetry(ctx, job, "record-step failed")
			return
		}

		if err := r.Repo.MarkSuccess(ctx, job.ID, time.Now()); err != nil {
			r.Logger.Printf("job %s MarkSuccess error: %v", job.ID, err)
			return
		}

		if !inserted {
			r.Logger.Printf("job %s SUCCESS (idempotent replay) (%s)", job.ID, time.Since(start))
			return
		}

		r.Logger.Printf("job %s SUCCESS (%s)", job.ID, time.Since(start))
		return
	}

	// failure path
	nextAttempts := job.Attempts + 1
	terminal := nextAttempts >= job.MaxAttempts

	if terminal {
		err := r.Repo.MarkFailure(ctx, job.ID, nextAttempts, nil, "simulated failure", true, ptrTime(time.Now()))
		if err != nil {
			r.Logger.Printf("job %s MarkFailure(terminal) error: %v", job.ID, err)
			return
		}
		r.Logger.Printf("job %s FAILED terminal attempts=%d/%d", job.ID, nextAttempts, job.MaxAttempts)
		return
	}

	delay := r.Backoff.Next(nextAttempts)
	nextRun := time.Now().Add(delay)

	err := r.Repo.MarkFailure(ctx, job.ID, nextAttempts, &nextRun, "simulated failure", false, nil)
	if err != nil {
		r.Logger.Printf("job %s MarkFailure(retry) error: %v", job.ID, err)
		return
	}
	r.Logger.Printf("job %s RETRY scheduled attempts=%d/%d next_in=%s", job.ID, nextAttempts, job.MaxAttempts, delay)
}

func (r *Runner) scheduleRetry(ctx context.Context, job Job, msg string) {
	nextAttempts := job.Attempts + 1
	terminal := nextAttempts >= job.MaxAttempts

	if terminal {
		_ = r.Repo.MarkFailure(ctx, job.ID, nextAttempts, nil, msg, true, ptrTime(time.Now()))
		return
	}

	delay := r.Backoff.Next(nextAttempts)
	nextRun := time.Now().Add(delay)
	_ = r.Repo.MarkFailure(ctx, job.ID, nextAttempts, &nextRun, msg, false, nil)
}

func ptrTime(t time.Time) *time.Time { return &t }
