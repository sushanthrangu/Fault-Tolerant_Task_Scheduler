package mysqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"task-scheduler/internal/domain"
)

type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

/*
====================================================
API METHODS
====================================================
*/

func (r *JobRepo) CreateJob(
	ctx context.Context,
	id string,
	jobType string,
	payload []byte,
	maxAttempts int,
	idempotencyKey *string,
) (*domain.Job, error) {

	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if jobType == "" {
		return nil, fmt.Errorf("jobType is required")
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload is required")
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO jobs (
			id, type, payload, status,
			attempts, max_attempts,
			next_run_at,
			idempotency_key
		) VALUES (
			?, ?, ?, 'PENDING',
			0, ?,
			NOW(6),
			?
		)
	`, id, jobType, payload, maxAttempts, idempotencyKey)

	if err != nil {
		if idempotencyKey != nil {
			existing, getErr := r.GetJobByIdempotencyKey(ctx, *idempotencyKey)
			if getErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("insert job: %w", err)
	}

	return r.GetJobByID(ctx, id)
}

func (r *JobRepo) GetJobByID(ctx context.Context, id string) (*domain.Job, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id, type, CAST(payload AS CHAR),
			status, attempts, max_attempts,
			next_run_at,
			idempotency_key,
			started_at, completed_at, error_message,
			locked_by, locked_until,
			created_at, updated_at
		FROM jobs
		WHERE id = ?
	`, id)

	return scanJob(row)
}

func (r *JobRepo) GetJobByIdempotencyKey(ctx context.Context, key string) (*domain.Job, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id, type, CAST(payload AS CHAR),
			status, attempts, max_attempts,
			next_run_at,
			idempotency_key,
			started_at, completed_at, error_message,
			locked_by, locked_until,
			created_at, updated_at
		FROM jobs
		WHERE idempotency_key = ?
		LIMIT 1
	`, key)

	return scanJob(row)
}

/*
====================================================
WORKER METHODS (TEMP STUBS)
====================================================
*/

func (r *JobRepo) ClaimJobs(
	ctx context.Context,
	workerID string,
	limit int,
	lease time.Duration,
	now time.Time,
) ([]domain.Job, error) {

	if workerID == "" {
		return nil, fmt.Errorf("workerID required")
	}
	if limit <= 0 {
		limit = 10
	}

	leaseUntil := now.Add(lease)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM jobs
		WHERE
			(
				status = 'PENDING'
				AND (next_run_at IS NULL OR next_run_at <= ?)
				AND (locked_until IS NULL OR locked_until <= ?)
			)
			OR
			(
				status = 'RUNNING'
				AND locked_until IS NOT NULL
				AND locked_until <= ?
			)
		ORDER BY next_run_at ASC
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`, now, now, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		tx.Commit()
		return []domain.Job{}, nil
	}

	// Update them to RUNNING + lease
	for _, id := range ids {
		_, err := tx.ExecContext(ctx, `
			UPDATE jobs
			SET
				status = 'RUNNING',
				locked_by = ?,
				locked_until = ?,
				started_at = COALESCE(started_at, ?)
			WHERE id = ?
		`, workerID, leaseUntil, now, id)
		if err != nil {
			return nil, err
		}
	}

	// Fetch full rows
	var claimed []domain.Job
	for _, id := range ids {
		row := tx.QueryRowContext(ctx, `
			SELECT
				id, type, CAST(payload AS CHAR),
				status, attempts, max_attempts,
				next_run_at,
				idempotency_key,
				started_at, completed_at, error_message,
				locked_by, locked_until,
				created_at, updated_at
			FROM jobs
			WHERE id = ?
		`, id)

		job, err := scanJob(row)
		if err != nil {
			return nil, err
		}
		if job != nil {
			claimed = append(claimed, *job)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return claimed, nil
}

func (r *JobRepo) Heartbeat(
	ctx context.Context,
	jobID string,
	workerID string,
	extendBy time.Duration,
	now time.Time,
) error {
	return fmt.Errorf("Heartbeat not implemented yet")
}

func (r *JobRepo) MarkSuccess(
	ctx context.Context,
	jobID string,
	completedAt time.Time,
) error {
	if jobID == "" {
		return fmt.Errorf("jobID is required")
	}
	if completedAt.IsZero() {
		completedAt = time.Now()
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE jobs
		SET
			status = 'SUCCESS',
			completed_at = ?,
			error_message = NULL,
			locked_by = NULL,
			locked_until = NULL
		WHERE id = ? AND status = 'RUNNING'
	`, completedAt, jobID)
	if err != nil {
		return fmt.Errorf("mark success update: %w", err)
	}

	aff, _ := res.RowsAffected()
	if aff == 0 {
		return fmt.Errorf("mark success rejected: job not RUNNING or not found")
	}
	return nil
}

func (r *JobRepo) MarkFailure(
	ctx context.Context,
	jobID string,
	attempts int,
	nextRunAt *time.Time,
	errMsg string,
	terminal bool,
	completedAt *time.Time,
) error {
	if jobID == "" {
		return fmt.Errorf("jobID is required")
	}
	if errMsg == "" {
		errMsg = "unknown error"
	}

	status := "PENDING"
	var comp any = nil
	if terminal {
		status = "FAILED"
		// terminal failures should have completed_at
		if completedAt != nil {
			comp = *completedAt
		} else {
			comp = time.Now()
		}
	}

	// For retry: keep completed_at NULL and set next_run_at
	var next any = nil
	if !terminal && nextRunAt != nil {
		next = *nextRunAt
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE jobs
		SET
			status = ?,
			attempts = ?,
			next_run_at = ?,
			completed_at = ?,
			error_message = ?,
			locked_by = NULL,
			locked_until = NULL
		WHERE id = ? AND status = 'RUNNING'
	`, status, attempts, next, comp, errMsg, jobID)
	if err != nil {
		return fmt.Errorf("mark failure update: %w", err)
	}

	aff, _ := res.RowsAffected()
	if aff == 0 {
		return fmt.Errorf("mark failure rejected: job not RUNNING or not found")
	}
	return nil
}

func (r *JobRepo) RecordStepOnce(
	ctx context.Context,
	jobID string,
	stepKey string,
	resultHash *string,
) (bool, error) {
	if jobID == "" {
		return false, fmt.Errorf("jobID is required")
	}
	if stepKey == "" {
		return false, fmt.Errorf("stepKey is required")
	}

	// Insert once by PK(job_id, step_key). If duplicate => already executed.
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO job_executions (job_id, step_key, result_hash)
		VALUES (?, ?, ?)
	`, jobID, stepKey, resultHash)

	if err != nil {
		// MySQL duplicate key error: treat as "already executed"
		// We keep this lightweight without driver-specific error types.
		if isDuplicateKey(err) {
			return false, nil
		}
		return false, fmt.Errorf("record step: %w", err)
	}

	return true, nil
}

func isDuplicateKey(err error) bool {
	// Works across common MySQL drivers because message includes "Duplicate entry".
	// Keeps us dependency-light.
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}

/*
====================================================
HELPERS
====================================================
*/

type jobRow interface {
	Scan(dest ...any) error
}

func scanJob(row jobRow) (*domain.Job, error) {
	var j domain.Job
	var payloadStr string

	var nextRunAt sql.NullTime
	var idemKey sql.NullString
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var errMsg sql.NullString
	var lockedBy sql.NullString
	var lockedUntil sql.NullTime

	err := row.Scan(
		&j.ID, &j.Type, &payloadStr,
		&j.Status, &j.Attempts, &j.MaxAttempts,
		&nextRunAt,
		&idemKey,
		&startedAt, &completedAt, &errMsg,
		&lockedBy, &lockedUntil,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	j.Payload = []byte(payloadStr)

	if nextRunAt.Valid {
		t := nextRunAt.Time
		j.NextRunAt = &t
	}
	if idemKey.Valid {
		s := idemKey.String
		j.IdempotencyKey = &s
	}
	if startedAt.Valid {
		t := startedAt.Time
		j.StartedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		j.CompletedAt = &t
	}
	if errMsg.Valid {
		s := errMsg.String
		j.ErrorMessage = &s
	}
	if lockedBy.Valid {
		s := lockedBy.String
		j.LockedBy = &s
	}
	if lockedUntil.Valid {
		t := lockedUntil.Time
		j.LockedUntil = &t
	}

	return &j, nil
}
