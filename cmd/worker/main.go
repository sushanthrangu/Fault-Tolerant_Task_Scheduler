package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"task-scheduler/internal/config"
	mysqlrepo "task-scheduler/internal/repo/mysql"
	"task-scheduler/internal/worker"
)

func main() {
	cfg := config.Load()
	if cfg.DBDSN == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := mysqlrepo.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	defer db.Close()

	repo := mysqlrepo.NewJobRepo(db)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("shutting down worker...")
		cancel()
	}()

	// Env tuning
	failRate := envFloat("FAIL_RATE", 0.30)
	poolSize := envInt("WORKER_POOL_SIZE", 4)
	queueSize := envInt("JOB_QUEUE_SIZE", 100)

	baseMs := envInt("BACKOFF_BASE_MS", 500)
	maxMs := envInt("BACKOFF_MAX_MS", 30_000)
	jitter := envFloat("BACKOFF_JITTER", 0.20)

	rand.Seed(time.Now().UnixNano())

	backoff := worker.BackoffConfig{
		Base:   time.Duration(baseMs) * time.Millisecond,
		Max:    time.Duration(maxMs) * time.Millisecond,
		Jitter: jitter,
	}

	runner := worker.NewRunner(repo, backoff, failRate, log.Default())
	pool := worker.NewPool(rootCtx, runner, poolSize, queueSize)

	log.Printf("worker started id=%s poll=%s pool=%d queue=%d fail_rate=%.2f",
		cfg.WorkerID, cfg.PollInterval, poolSize, queueSize, failRate,
	)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rootCtx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = pool.Stop(ctx)
			log.Println("worker stopped")
			return

		case <-ticker.C:
			now := time.Now()

			claimed, err := repo.ClaimJobs(
				rootCtx,
				cfg.WorkerID,
				10,
				time.Duration(cfg.LeaseSeconds)*time.Second,
				now,
			)
			if err != nil {
				log.Printf("claim error: %v", err)
				continue
			}

			for _, j := range claimed {
				ok := pool.Submit(worker.Job{
					ID:          j.ID,
					Attempts:    j.Attempts,
					MaxAttempts: j.MaxAttempts,
				})

				if !ok {
					// Backpressure: reschedule quickly and release lease.
					next := time.Now().Add(250 * time.Millisecond)
					_ = repo.MarkFailure(rootCtx, j.ID, j.Attempts, &next, "queue full - rescheduled", false, nil)
					log.Printf("queue full: rescheduled job %s", j.ID)
				}
			}
		}
	}
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func envFloat(k string, def float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}
