package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"task-scheduler/internal/domain"
	"task-scheduler/internal/repo"
)

type JobService struct {
	Repo repo.JobRepository
}

func NewJobService(r repo.JobRepository) *JobService {
	return &JobService{Repo: r}
}

func (s *JobService) Create(ctx context.Context, id, jobType string, payload json.RawMessage, maxAttempts int, idemKey *string) (*domain.Job, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id required", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(jobType) == "" {
		return nil, fmt.Errorf("%w: type required", domain.ErrInvalidInput)
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("%w: payload required", domain.ErrInvalidInput)
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return s.Repo.CreateJob(ctx, id, jobType, payload, maxAttempts, idemKey)
}

func (s *JobService) Get(ctx context.Context, id string) (*domain.Job, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id required", domain.ErrInvalidInput)
	}
	return s.Repo.GetJobByID(ctx, id)
}
