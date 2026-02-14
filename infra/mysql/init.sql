CREATE TABLE IF NOT EXISTS jobs (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    payload JSON NOT NULL,

    status ENUM('PENDING', 'RUNNING', 'SUCCESS', 'FAILED') 
        NOT NULL DEFAULT 'PENDING',

    -- Retry mechanism
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    next_run_at TIMESTAMP NULL,

    -- Idempotency
    idempotency_key VARCHAR(255) UNIQUE,

    -- Distributed locking (CRITICAL)
    locked_by VARCHAR(64) NULL,
    locked_until TIMESTAMP NULL,

    -- Execution tracking
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error_message TEXT,

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
        ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_pick (status, next_run_at, locked_until),
    INDEX idx_idempotency_key (idempotency_key),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- Exactly-once execution guard per job step
CREATE TABLE IF NOT EXISTS job_executions (
    job_id VARCHAR(36) NOT NULL,
    step_key VARCHAR(100) NOT NULL,
    result_hash VARCHAR(64) NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (job_id, step_key),
    INDEX idx_executed_at (executed_at),

    CONSTRAINT fk_job_exec_job
      FOREIGN KEY (job_id) REFERENCES jobs(id)
      ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
