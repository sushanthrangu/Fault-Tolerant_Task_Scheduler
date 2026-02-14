Fault-Tolerant Task Scheduler (Go + MySQL + Docker)

Production-style distributed task scheduler demonstrating:

REST API for job submission

Worker pool with goroutines

Exponential backoff retries

Lease-based distributed locking

Exactly-once execution guard

Graceful shutdown

Full Docker Compose stack

Architecture
Client
   │
   ▼
┌──────────┐      ┌──────────┐
│   API    │ ---> │  MySQL   │
└──────────┘      └──────────┘
                        ▲
                        │
                ┌─────────────┐
                │   Worker    │
                │ (goroutine  │
                │   pool)     │
                └─────────────┘

Core Guarantees
Feature	Implementation
Retry	Exponential backoff with jitter
Fault tolerance	Lease-based locking (locked_until)
Idempotency	idempotency_key + job_executions table
Exactly-once guard	(job_id, step_key) primary key
Horizontal scaling	Multiple workers safe via DB row locking
Backpressure	Bounded job queue
Quick Start (1 Command)

From deploy/:

docker compose up --build


Services:

Service	Host Port
API	http://localhost:8086

MySQL	3310

Health check:

curl http://localhost:8086/healthz

Create a Job
curl -X POST http://localhost:8086/jobs \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-1" \
  -d '{
        "type": "demo",
        "payload": {"msg": "hello"},
        "max_attempts": 3
      }'

Fetch Job Status
curl http://localhost:8086/jobs/<job_id>


Response Example:

{
  "id": "7dab128f237ee00ce17961b4963b31b7",
  "type": "demo",
  "status": "SUCCESS",
  "attempts": 1,
  "max_attempts": 3,
  "payload": {
    "msg": "hello"
  },
  "created_at": "...",
  "updated_at": "..."
}

Retry Strategy

Exponential backoff with jitter:

delay = base * 2^(attempt-1)


Capped at configurable max.

Configurable via env:

BACKOFF_BASE_MS
BACKOFF_MAX_MS
BACKOFF_JITTER

Distributed Locking

Each job uses lease-based locking:

locked_by

locked_until

Workers only claim jobs where:

locked_until IS NULL OR locked_until < NOW()


Prevents double execution.

Exactly-Once Execution Guard

Table:

job_executions (job_id, step_key)


Before executing side-effect:

INSERT INTO job_executions


If duplicate → job already executed → mark SUCCESS without re-running side-effect.

Worker Pool

Configurable concurrency (WORKER_POOL_SIZE)

Bounded queue (JOB_QUEUE_SIZE)

Graceful drain on SIGINT/SIGTERM

Environment Variables
Variable	Purpose
DB_DSN	MySQL connection
PORT	API port
WORKER_ID	Unique worker name
POLL_INTERVAL_MS	Claim interval
LEASE_SECONDS	Lock lease duration
WORKER_POOL_SIZE	Concurrency
FAIL_RATE	Failure injection for testing
Local Development (without Docker)

Start MySQL:

docker compose up -d mysql


Run API:

export DB_DSN="root:root@tcp(127.0.0.1:3310)/scheduler?parseTime=true"
go run ./cmd/api


Run Worker:

export DB_DSN="root:root@tcp(127.0.0.1:3310)/scheduler?parseTime=true"
go run ./cmd/worker

Scaling

Run multiple workers:

docker compose up --scale worker=3


Safe due to:

Row-level locking

Lease expiry

Idempotent execution

Future Enhancements

Dead Letter Queue table

Prometheus metrics

OpenTelemetry tracing

Cron-style scheduled jobs

Redis-based rate limiting

Circuit breaker for downstream calls

Multi-step workflow engine

Why This Project Matters

This project demonstrates:

Distributed systems design

Concurrency control

Fault tolerance patterns

Idempotency design

Production-grade retry logic

Clean architecture in Go