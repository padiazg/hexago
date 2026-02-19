# hexago add worker

Add a background worker to an existing project.

## Synopsis

```shell
hexago add worker <name> [flags]
```

Must be run from the project root directory.

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | *(required)* | Worker type: `queue`, `periodic`, or `event` |
| `--interval` | `-i` | string | `""` | Execution interval for periodic workers (e.g. `5m`, `1h`, `30s`) |
| `--workers` | `-w` | int | `1` | Number of goroutines for queue workers |
| `--queue-size` | `-q` | int | `100` | Job queue buffer size for queue workers |

---

## Worker Types

### Queue Worker

A worker pool that processes jobs from a buffered channel.

```shell
hexago add worker EmailWorker --type queue --workers 5
hexago add worker ImageProcessor --type queue --workers 3 --queue-size 500
```

Use for: processing tasks asynchronously, handling message queue messages, parallel job execution.

### Periodic Worker

Executes a task on a fixed schedule.

```shell
hexago add worker HealthWorker --type periodic --interval 5m
hexago add worker CacheWarmer --type periodic --interval 10m
hexago add worker ReportGenerator --type periodic --interval 1h
```

Use for: scheduled tasks, cache warming, periodic report generation, cleanup jobs.

### Event Worker

Reacts to events published on a channel.

```shell
hexago add worker NotificationWorker --type event
hexago add worker AuditLogWorker --type event
```

Use for: event-driven processing, audit logging, reactive workflows.

---

## Examples

```shell
# Queue-based worker pool with 5 goroutines
hexago add worker EmailWorker --type queue --workers 5

# Periodic health check every 5 minutes
hexago add worker HealthWorker --type periodic --interval 5m

# Event-driven notification processor
hexago add worker NotificationWorker --type event

# Large queue for image processing
hexago add worker ImageProcessor --type queue --workers 10 --queue-size 1000
```

---

## Generated Files

For `hexago add worker EmailWorker --type queue`:

```
internal/
└── workers/
    └── email_worker.go
```

---

## Generated Code Structure

**Queue Worker:**

```go
package workers

import (
    "context"
    "log"
)

// EmailWorkerJob represents a job for EmailWorker
type EmailWorkerJob struct {
    // TODO: Add job fields
}

// EmailWorker processes jobs from a queue
type EmailWorker struct {
    jobQueue chan EmailWorkerJob
    workers  int
}

// NewEmailWorker creates a new EmailWorker
func NewEmailWorker(workers int, queueSize int) *EmailWorker {
    return &EmailWorker{
        jobQueue: make(chan EmailWorkerJob, queueSize),
        workers:  workers,
    }
}

// Start starts the worker pool
func (w *EmailWorker) Start(ctx context.Context) {
    for i := 0; i < w.workers; i++ {
        go w.runWorker(ctx)
    }
}

// Submit adds a job to the queue
func (w *EmailWorker) Submit(job EmailWorkerJob) {
    w.jobQueue <- job
}

func (w *EmailWorker) runWorker(ctx context.Context) {
    for {
        select {
        case job := <-w.jobQueue:
            if err := w.processJob(ctx, job); err != nil {
                log.Printf("EmailWorker error: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

func (w *EmailWorker) processJob(ctx context.Context, job EmailWorkerJob) error {
    // TODO: Implement job processing
    return nil
}
```
