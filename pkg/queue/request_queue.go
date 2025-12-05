package queue

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	// ErrQueueFull is returned when queue is at capacity
	ErrQueueFull = errors.New("request queue is full")
	
	// ErrQueueClosed is returned when queue is closed
	ErrQueueClosed = errors.New("request queue is closed")
)

// Request represents a queued request
type Request struct {
	ID        string
	Data      interface{}
	Context   context.Context
	Result    chan Result
	EnqueueAt time.Time
}

// Result represents the result of a processed request
type Result struct {
	Data  interface{}
	Error error
}

// RequestProcessor processes queued requests
type RequestProcessor func(ctx context.Context, data interface{}) (interface{}, error)

// RequestQueue implements a buffered channel-based request queue
// with worker pool for processing
type RequestQueue struct {
	queue     chan *Request
	processor RequestProcessor
	workers   int
	
	mu        sync.RWMutex
	wg        sync.WaitGroup
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
	
	// Metrics
	totalEnqueued   int64
	totalProcessed  int64
	totalFailed     int64
	totalDropped    int64
}

// Config holds queue configuration
type Config struct {
	QueueSize int              // Buffer size for queue
	Workers   int              // Number of worker goroutines
	Processor RequestProcessor // Function to process requests
}

// NewRequestQueue creates a new request queue
func NewRequestQueue(config Config) *RequestQueue {
	if config.QueueSize <= 0 {
		config.QueueSize = 1000
	}
	if config.Workers <= 0 {
		config.Workers = 10
	}
	if config.Processor == nil {
		panic("processor function is required")
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RequestQueue{
		queue:     make(chan *Request, config.QueueSize),
		processor: config.Processor,
		workers:   config.Workers,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the worker pool
func (q *RequestQueue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.closed {
		log.Printf("⚠️ Cannot start closed queue")
		return
	}
	
	log.Printf("🚀 Starting request queue with %d workers", q.workers)
	
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// worker processes requests from the queue
func (q *RequestQueue) worker(id int) {
	defer q.wg.Done()
	
	log.Printf("👷 Worker %d started", id)
	
	for {
		select {
		case <-q.ctx.Done():
			log.Printf("👷 Worker %d stopped", id)
			return
			
		case req, ok := <-q.queue:
			if !ok {
				log.Printf("👷 Worker %d: queue closed", id)
				return
			}
			
			q.processRequest(id, req)
		}
	}
}

// processRequest processes a single request
func (q *RequestQueue) processRequest(workerID int, req *Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Worker %d: panic processing request %s: %v", workerID, req.ID, r)
			q.mu.Lock()
			q.totalFailed++
			q.mu.Unlock()
			
			req.Result <- Result{
				Error: errors.New("internal error processing request"),
			}
		}
	}()
	
	waitTime := time.Since(req.EnqueueAt)
	log.Printf("⚙️ Worker %d: processing request %s (waited %v)", workerID, req.ID, waitTime)
	
	// Check if request context is still valid
	if req.Context.Err() != nil {
		log.Printf("⚠️ Worker %d: request %s context cancelled", workerID, req.ID)
		q.mu.Lock()
		q.totalFailed++
		q.mu.Unlock()
		
		req.Result <- Result{
			Error: req.Context.Err(),
		}
		return
	}
	
	// Process the request
	result, err := q.processor(req.Context, req.Data)
	
	q.mu.Lock()
	if err != nil {
		q.totalFailed++
	} else {
		q.totalProcessed++
	}
	q.mu.Unlock()
	
	// Send result back
	select {
	case req.Result <- Result{Data: result, Error: err}:
		log.Printf("✅ Worker %d: completed request %s", workerID, req.ID)
	case <-time.After(5 * time.Second):
		log.Printf("⚠️ Worker %d: timeout sending result for request %s", workerID, req.ID)
	}
}

// Enqueue adds a request to the queue (non-blocking)
func (q *RequestQueue) Enqueue(req *Request) error {
	q.mu.RLock()
	if q.closed {
		q.mu.RUnlock()
		return ErrQueueClosed
	}
	q.mu.RUnlock()
	
	// Try to enqueue without blocking
	select {
	case q.queue <- req:
		q.mu.Lock()
		q.totalEnqueued++
		q.mu.Unlock()
		
		log.Printf("📥 Request %s enqueued (queue size: %d/%d)", 
			req.ID, len(q.queue), cap(q.queue))
		return nil
		
	default:
		// Queue is full
		q.mu.Lock()
		q.totalDropped++
		q.mu.Unlock()
		
		log.Printf("❌ Request %s dropped - queue full (%d/%d)", 
			req.ID, len(q.queue), cap(q.queue))
		return ErrQueueFull
	}
}

// Stop gracefully stops the queue and waits for workers to finish
func (q *RequestQueue) Stop(timeout time.Duration) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	q.mu.Unlock()
	
	log.Printf("🛑 Stopping request queue...")
	
	// Close queue channel (no more enqueues)
	close(q.queue)
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Printf("✅ Request queue stopped gracefully")
		return nil
	case <-time.After(timeout):
		log.Printf("⚠️ Request queue stop timeout - cancelling workers")
		q.cancel()
		return errors.New("queue stop timeout")
	}
}

// Stats returns queue statistics
func (q *RequestQueue) Stats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	return map[string]interface{}{
		"queue_size":      len(q.queue),
		"queue_capacity":  cap(q.queue),
		"workers":         q.workers,
		"total_enqueued":  q.totalEnqueued,
		"total_processed": q.totalProcessed,
		"total_failed":    q.totalFailed,
		"total_dropped":   q.totalDropped,
		"closed":          q.closed,
	}
}

// QueueSize returns current queue size
func (q *RequestQueue) QueueSize() int {
	return len(q.queue)
}

// IsFull returns true if queue is at capacity
func (q *RequestQueue) IsFull() bool {
	return len(q.queue) >= cap(q.queue)
}
