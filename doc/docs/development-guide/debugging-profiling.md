# Debugging & Profiling

Tools and techniques for debugging and profiling Go applications.

---

## Debugging

### Delve Debugger

Install Delve:

```shell
go install github.com/go-delve/delve/cmd/dlv@latest
```

Start debugging:

```shell
# Debug the entire package
dlv debug ./...

# Debug a specific file
dlv debug ./cmd/server.go

# Attach to running process
dlv attach <pid>
```

### Common Delve Commands

| Command | Shortcut | Description |
| --- | --- | --- |
| `break main.go:10` | `b` | Set breakpoint |
| `break foo.go:15` | | Set breakpoint at line |
| `continue` | `c` | Run until breakpoint |
| `next` | `n` | Step over line |
| `step` | `s` | Step into function |
| `out` | `o` | Step out of function |
| `print variable` | `p` | Print variable value |
| `locals` | | Print local variables |
| `stack` | `bt` | Print stack trace |
| `exit` | `q` | Quit debugger |

### Example Debug Session

```shell
(dlv) break internal/core/services/user.go:25
(dlv) continue
> internal/core/services/user.go:25
user.go:25          return s.store.GetUser(ctx, id)
(dlv) print id
"123"
(dlv) locals
ctx = context.Background
id = "123"
(dlv) continue
```

---

## Profiling

### CPU Profiling

```shell
# Run with CPU profiling enabled
go test -cpuprofile=cpu.prof -bench=.

# Or for the application
go run -cpuprofile=cpu.prof main.go run
```

Analyze the profile:

```shell
go tool pprof cpu.prof

# Interactive commands:
(pprof) top
(pprof) top25
(pprof) list FunctionName
(pprof) web
```

### Memory Profiling

```shell
# Run with memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze
go tool pprof mem.prof

(pprof) top
(pprof) top -cum
(pprof) list FunctionName
```

### Block Profiling

Identifies synchronization bottlenecks:

```shell
go test -blockprofile=block.prof -bench=.
go tool pprof block.prof
```

### Mutex Profiling

Find lock contention:

```shell
go test -mutexprofile=mutex.prof -bench=.
go tool pprof mutex.prof
```

---

## Benchmarks

### Running Benchmarks

```shell
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=^BenchmarkUserCreate$ -benchmem ./...

# Run with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.out ./...
```

### Benchmark Example

```go
func BenchmarkUserService_Create(b *testing.B) {
    store := NewMockStore()
    service := NewUserService(store)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.Create(context.Background(), &CreateRequest{
            Email: "test@example.com",
        })
    }
}
```

### Interpreting Results

```
BenchmarkUserService_Create-8    1000000    1234 ns/op    256 B/op    3 allocs/op
          ↑                         ↑          ↑             ↑           ↑
        Name                    Iterations   Time/op       Memory     Allocations
```

### Comparing Benchmarks

```shell
# Save baseline
go test -bench=. -count=5 > baseline.txt

# After changes
go test -bench=. -count=5 > new.txt

# Compare with benchstat
go install golang.org/x/tools/cmd/benchstat@latest
benchstat baseline.txt new.txt
```

---

## Tracing

### Generate Trace

```shell
go test -trace=trace.out ./...
go run -trace=trace.out main.go run
```

### View Trace

```shell
go tool trace trace.out
```

Trace shows:

- Goroutine scheduling
- Syscall blocking
- GC events
- User-defined tasks

---

## pprof Web Interface

Generate a profile and view in browser:

```shell
# Start HTTP server for pprof
go tool pprof -http=:8080 cpu.prof
```

This opens a browser with:

- Graph visualization
- Source code view
- Flame graph
- Call tree

---

## Common Performance Issues

| Issue | Profiling Type | Solution |
| --- | --- | --- |
| High CPU | CPU profile | Optimize hot functions |
| Memory leaks | Memory profile | Check retained objects |
| Slow requests | CPU + trace | Find blocking operations |
| Lock contention | Mutex profile | Reduce critical sections |
| Excessive allocs | Mem profile | Reduce allocations |

---

## Runtime Metrics

### GODEBUG Settings

```shell
# Show GC trace
GODEBUG=gctrace=1 go run main.go

# Show scheduler trace
GODEBUG=schedtrace=1000 go run main.go
```

### Prometheus Metrics

If observability is enabled:

```shell
# View metrics
curl http://localhost:9090/metrics

# Common metrics:
# - go_goroutines
# - go_memstats_alloc_bytes
# - process_cpu_seconds_total
```

---

## Integration with IDE

### VS Code

Install Go extension, then:

1. Set breakpoints in code  
2. Press F5 to start debugging  
3. Use Debug Console to inspect variables  

### GoLand

1. Run → Edit Configurations  
2. Add Go Remote or Go Test configuration  
3. Set breakpoints and run  

---

## Quick Reference

| Task | Command |
| --- | --- |
| Debug a test | `dlv test ./...` |
| CPU profile | `go test -cpuprofile=pprof.out` |
| Memory profile | `go test -memprofile=pprof.out` |
| Benchmarks | `go test -bench=. -benchmem` |
| View profile | `go tool pprof pprof.out` |
| Generate trace | `go test -trace=trace.out` |
| View trace | `go tool trace trace.out` |
