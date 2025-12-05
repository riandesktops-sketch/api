# Rate Limiting & Concurrent Request Handling

## Problem

Ketika mengirim 100+ request secara bersamaan ke Gemini API, terjadi error karena rate limiting:
- Free tier: ~15 requests per minute
- Paid tier: ~60 requests per minute
- Concurrent request limits

## Solution

Implementasi sistem 3-layer untuk menangani concurrent requests:

### 1. Rate Limiter (Token Bucket)
- **Location**: `pkg/ratelimiter/rate_limiter.go`
- **Algorithm**: Token Bucket dengan `golang.org/x/time/rate`
- **Default**: 10 requests per second (600 RPM)
- **Features**:
  - Thread-safe concurrent access
  - Dynamic rate adjustment
  - Non-blocking checks
  - Context-aware waiting

### 2. Request Queue (Worker Pool)
- **Location**: `pkg/queue/request_queue.go`
- **Pattern**: Buffered channel-based queue with worker pool
- **Default**: 1000 buffer size, 10 workers
- **Features**:
  - Non-blocking enqueue
  - Graceful shutdown
  - Request timeout handling
  - Real-time metrics

### 3. Circuit Breaker
- **Location**: `pkg/circuitbreaker/circuit_breaker.go`
- **Pattern**: Circuit Breaker with 3 states (Closed, Open, Half-Open)
- **Default**: 5 failures threshold, 60s timeout
- **Features**:
  - Fail-fast when API down
  - Automatic recovery testing
  - Prevents cascade failures

## Architecture

```
Client Request
    ↓
AI Handler (enqueue)
    ↓
Request Queue (1000 buffer)
    ↓
Worker Pool (10 workers)
    ↓
Rate Limiter (10 req/s)
    ↓
Circuit Breaker (5 failures)
    ↓
Gemini API
```

## Files Modified

### New Files
- `pkg/ratelimiter/rate_limiter.go` - Rate limiting implementation
- `pkg/circuitbreaker/circuit_breaker.go` - Circuit breaker pattern
- `pkg/queue/request_queue.go` - Request queue with worker pool
- `test/load_test.go` - Load testing script

### Modified Files
- `services/ai-service/client/gemini.go` - Added rate limiter & circuit breaker
- `services/ai-service/handlers/ai_handler.go` - Queue-based request handling
- `services/ai-service/main.go` - Queue initialization & lifecycle

## Usage

### Starting the Service

```bash
# Set Gemini API key
export GEMINI_API_KEY="your-api-key"

# Build and run
go build -o ai-service ./services/ai-service
./ai-service
```

### Load Testing

```bash
# Test with 10 concurrent requests
go run test/load_test.go -concurrent=10 -requests=100

# Test with 100 concurrent requests
go run test/load_test.go -concurrent=100 -requests=1000

# Test with 500 concurrent requests
go run test/load_test.go -concurrent=500 -requests=5000
```

### Monitoring

```bash
# Check health and queue stats
curl http://localhost:8084/health

# Response includes queue metrics:
{
  "status": "healthy",
  "service": "ai-service",
  "queue": {
    "queue_size": 45,
    "queue_capacity": 1000,
    "workers": 10,
    "total_enqueued": 1523,
    "total_processed": 1478,
    "total_failed": 12,
    "total_dropped": 8
  }
}
```

## Configuration

### Rate Limiter

```go
// In gemini.go NewGeminiClient()

// Free tier (15 RPM)
rateLimiter := ratelimiter.NewRateLimiter(15, 60*time.Second)

// Paid tier (60 RPM)
rateLimiter := ratelimiter.NewRateLimiter(60, 60*time.Second)

// High volume (600 RPM)
rateLimiter := ratelimiter.NewRateLimiter(10, time.Second)
```

### Request Queue

```go
// In main.go

requestQueue := queue.NewRequestQueue(queue.Config{
    QueueSize: 1000,  // Buffer size
    Workers:   10,    // Concurrent workers
    Processor: processorFunc,
})
```

### Circuit Breaker

```go
// In gemini.go NewGeminiClient()

circuitBreaker := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
    MaxFailures:     5,              // Failures before opening
    ResetTimeout:    60 * time.Second, // Time before retry
    HalfOpenMaxReqs: 1,              // Test requests in half-open
})
```

## Error Handling

### Queue Full (429)
When queue is at capacity, returns:
```json
{
  "success": false,
  "message": "Server is busy, please try again later"
}
```

**Client should retry with exponential backoff**

### Circuit Breaker Open
When circuit is open, requests fail fast without calling API:
```json
{
  "success": false,
  "message": "Failed to generate AI response"
}
```

**Fallback response is returned automatically**

### Request Timeout
After 60 seconds waiting in queue:
```json
{
  "success": false,
  "message": "Request timeout - please try again"
}
```

## Performance

### Capacity
- **Queue Buffer**: 1000 requests
- **Concurrent Workers**: 10
- **Rate Limit**: 10 req/s (600 RPM)
- **Max Throughput**: ~6-8 req/s (sustained)

### Expected Results
- ✅ Success rate: >= 95%
- ✅ Average latency: < 5 seconds
- ✅ Queue full errors: < 5% (extreme load)
- ✅ No Gemini API rate limit errors

## Scaling

### Vertical Scaling
Increase workers and queue size:
```go
requestQueue := queue.NewRequestQueue(queue.Config{
    QueueSize: 5000,  // Larger buffer
    Workers:   50,    // More workers
    Processor: processorFunc,
})
```

### Horizontal Scaling
Deploy multiple instances behind load balancer:
```
Load Balancer
    ├── AI Service Instance 1
    ├── AI Service Instance 2
    └── AI Service Instance 3
```

## Monitoring & Alerts

### Key Metrics
- `queue_size` - Current requests in queue
- `total_enqueued` - Total requests received
- `total_processed` - Successfully processed
- `total_failed` - Failed requests
- `total_dropped` - Rejected (queue full)

### Recommended Alerts
- Queue size > 80% capacity
- Circuit breaker opens
- Success rate < 95%
- Average latency > 10 seconds

## Troubleshooting

### High Queue Size
1. Increase worker count
2. Check circuit breaker state
3. Verify rate limiter settings

### Frequent 429 Errors
1. Increase queue buffer
2. Implement client-side rate limiting
3. Scale horizontally

### Circuit Breaker Stuck Open
1. Check API key validity
2. Verify network connectivity
3. Review error logs

## Dependencies

```bash
# Install required package
go get golang.org/x/time/rate
go mod tidy
```

## Testing Checklist

- [x] Build successfully
- [ ] Unit tests for rate limiter
- [ ] Unit tests for circuit breaker
- [ ] Unit tests for request queue
- [ ] Load test: 10 concurrent
- [ ] Load test: 100 concurrent
- [ ] Load test: 500 concurrent
- [ ] Integration test with real API
- [ ] Graceful shutdown test

## Next Steps

1. Run load tests to verify performance
2. Monitor metrics in production
3. Adjust configuration based on actual load
4. Add unit tests for new components
5. Consider adding request prioritization
6. Implement distributed rate limiting (Redis)

## References

- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Worker Pool Pattern](https://gobyexample.com/worker-pools)
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)
