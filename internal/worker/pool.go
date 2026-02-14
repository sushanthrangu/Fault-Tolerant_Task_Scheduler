package worker

import (
	"context"
	"sync"
)

type Job struct {
	ID          string
	Attempts    int
	MaxAttempts int
}

type Handler interface {
	Process(ctx context.Context, job Job)
}

// Pool is a bounded worker pool with graceful shutdown.
type Pool struct {
	ch     chan Job
	wg     sync.WaitGroup
	once   sync.Once
	closed chan struct{}

	ctx context.Context
	h   Handler
}

// NewPool starts `workers` goroutines reading from a bounded queue of size `queueSize`.
func NewPool(ctx context.Context, h Handler, workers int, queueSize int) *Pool {
	if ctx == nil {
		ctx = context.Background()
	}
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = 1
	}

	p := &Pool{
		ch:     make(chan Job, queueSize),
		closed: make(chan struct{}),
		ctx:    ctx,
		h:      h,
	}

	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer p.wg.Done()
			for job := range p.ch {
				h.Process(p.ctx, job)
			}
		}()
	}

	return p
}

// Submit tries to enqueue a job without blocking.
// Returns false if the queue is full or pool is stopped.
func (p *Pool) Submit(job Job) bool {
	select {
	case <-p.closed:
		return false
	default:
	}
	select {
	case p.ch <- job:
		return true
	default:
		return false
	}
}

// Stop stops accepting new work and drains existing queued work.
func (p *Pool) Stop(ctx context.Context) error {
	p.once.Do(func() {
		close(p.closed)
		close(p.ch)
	})

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
