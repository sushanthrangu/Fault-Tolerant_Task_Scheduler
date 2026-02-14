package domain

import (
	"encoding/json"
	"time"
)

type JobStatus string

const (
	StatusPending JobStatus = "PENDING"
	StatusRunning JobStatus = "RUNNING"
	StatusSuccess JobStatus = "SUCCESS"
	StatusFailed  JobStatus = "FAILED"
)

// Job is the canonical model used across API, service, repo, worker.
type Job struct {
	ID string `json:"id"`

	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"` // Prevent base64 encoding

	Status JobStatus `json:"status"`

	// Retry
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	NextRunAt   *time.Time `json:"next_run_at,omitempty"`

	// Idempotency
	IdempotencyKey *string `json:"idempotency_key,omitempty"`

	// Execution tracking
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`

	// Distributed locking (lease)
	LockedBy    *string    `json:"locked_by,omitempty"`
	LockedUntil *time.Time `json:"locked_until,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
