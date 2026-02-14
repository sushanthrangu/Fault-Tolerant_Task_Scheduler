# Fault-Tolerant Task Scheduler

A production-ready distributed task scheduler built with Go, MySQL, and Docker, demonstrating enterprise-grade patterns for reliable job processing.

## Features

- **REST API** for job submission and status tracking
- **Worker pool** with configurable goroutine concurrency
- **Exponential backoff** with jitter for intelligent retries
- **Lease-based distributed locking** for multi-worker safety
- **Exactly-once execution** guarantees via idempotency keys
- **Graceful shutdown** handling for zero job loss
- **Full Docker Compose** stack for one-command deployment

## Architecture

```
Client
   │
   ▼
┌──────────┐      ┌──────────┐
│   API    │ ───> │  MySQL   │
└──────────┘      └──────────┘
                        ▲
                        │
                ┌─────────────┐
                │   Worker    │
                │ (goroutine  │
                │   pool)     │
                └─────────────┘
```

## Core Guarantees

| Feature | Implementation |
|---------|----------------|
| **Retry** | Exponential backoff with jitter |
| **Fault tolerance** | Lease-based locking (`locked_until`) |
| **Idempotency** | `idempotency_key` + `job_executions` table |
| **Exactly-once guard** | `(job_id, step_key)` primary key |
| **Horizontal scaling** | Multiple workers safe via DB row locking |
| **Backpressure** | Bounded job queue |

---

## Quick Start

### Deploy with Docker Compose

From the `deploy/` directory:

```bash
docker compose up --build
```

**Services:**

| Service | Port | URL |
|---------|------|-----|
| API | 8086 | http://localhost:8086 |
| MySQL | 3310 | `root@localhost:3310` |

**Health check:**

```bash
curl http://localhost:8086/healthz
```

---

## API Usage

### Create a Job

```bash
curl -X POST http://localhost:8086/jobs \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-1" \
  -d '{
    "type": "demo",
    "payload": {"msg": "hello"},
    "max_attempts": 3
  }'
```

### Fetch Job Status

```bash
curl http://localhost:8086/jobs/<job_id>
```

**Response:**

```json
{
  "id": "7dab128f237ee00ce17961b4963b31b7",
  "type": "demo",
  "status": "SUCCESS",
  "attempts": 1,
  "max_attempts": 3,
  "payload": {
    "msg": "hello"
  },
  "created_at": "2026-02-14T10:30:00Z",
  "updated_at": "2026-02-14T10:30:05Z"
}
```

---

## Key Mechanisms

### Retry Strategy

Exponential backoff with jitter prevents thundering herd:

```
delay = base × 2^(attempt-1) + random_jitter
```

Configurable via environment variables:
- `BACKOFF_BASE_MS` – Initial delay (default: 1000ms)
- `BACKOFF_MAX_MS` – Maximum delay cap (default: 60000ms)
- `BACKOFF_JITTER` – Jitter factor (default: 0.1)

### Distributed Locking

Each job row includes:
- `locked_by` – Worker identifier
- `locked_until` – Lease expiration timestamp

Workers only claim jobs where:
```sql
locked_until IS NULL OR locked_until < NOW()
```

Prevents double execution across multiple workers.

### Exactly-Once Execution Guard

The `job_executions` table enforces:

```sql
PRIMARY KEY (job_id, step_key)
```

**Workflow:**
1. Before executing side-effects, insert `(job_id, step_key)`
2. If duplicate key error → job already executed
3. Mark job as `SUCCESS` without re-running side-effects

### Worker Pool

- Configurable concurrency (`WORKER_POOL_SIZE`)
- Bounded queue (`JOB_QUEUE_SIZE`) for backpressure
- Graceful drain on `SIGINT`/`SIGTERM`

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_DSN` | MySQL connection string | *required* |
| `PORT` | API server port | `8086` |
| `WORKER_ID` | Unique worker identifier | `worker-1` |
| `POLL_INTERVAL_MS` | Job claim polling interval | `5000` |
| `LEASE_SECONDS` | Lock lease duration | `30` |
| `WORKER_POOL_SIZE` | Concurrent goroutines | `10` |
| `JOB_QUEUE_SIZE` | Internal queue capacity | `100` |
| `BACKOFF_BASE_MS` | Initial retry delay | `1000` |
| `BACKOFF_MAX_MS` | Maximum retry delay | `60000` |
| `BACKOFF_JITTER` | Jitter randomization | `0.1` |
| `FAIL_RATE` | Failure injection (testing) | `0.0` |

---

## Local Development

### Without Docker

**1. Start MySQL:**

```bash
docker compose up -d mysql
```

**2. Run API:**

```bash
export DB_DSN="root:root@tcp(127.0.0.1:3310)/scheduler?parseTime=true"
go run ./cmd/api
```

**3. Run Worker:**

```bash
export DB_DSN="root:root@tcp(127.0.0.1:3310)/scheduler?parseTime=true"
export WORKER_ID="worker-local"
go run ./cmd/worker
```

---

## Scaling

Run multiple workers for horizontal scaling:

```bash
docker compose up --scale worker=3
```

**Safety guaranteed by:**
- Row-level database locking
- Lease expiration
- Idempotent execution guards

---

## Project Structure

```
.
├── cmd/
│   ├── api/          # HTTP server
│   └── worker/       # Job processor
├── deploy/
│   └── docker-compose.yml
├── internal/
│   ├── db/           # Database layer
│   ├── handler/      # HTTP handlers
│   ├── scheduler/    # Job scheduling logic
│   └── worker/       # Worker pool implementation
└── migrations/       # SQL schema
```

---

## Future Enhancements

- [ ] Dead Letter Queue (DLQ) for permanently failed jobs
- [ ] Prometheus metrics (`job_success_total`, `job_duration_seconds`)
- [ ] OpenTelemetry distributed tracing
- [ ] Cron-style scheduled jobs (`0 0 * * *`)
- [ ] Redis-based rate limiting
- [ ] Circuit breaker for downstream service calls
- [ ] Multi-step workflow engine (DAG support)
- [ ] Web UI dashboard for job monitoring

---

## Why This Project Matters

This scheduler demonstrates production-grade distributed systems patterns:

- **Distributed systems design** – Coordination without centralized state
- **Concurrency control** – Safe parallel job processing
- **Fault tolerance** – Graceful handling of crashes and network failures
- **Idempotency** – Safe retries without duplicate side-effects
- **Retry logic** – Exponential backoff prevents cascade failures
- **Clean architecture** – Separation of concerns in Go

Perfect for learning how to build resilient backend systems that scale.

---

## License

MIT

---

## Contributing

Contributions welcome! Please open an issue or PR.

**Development setup:**
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open a Pull Request
