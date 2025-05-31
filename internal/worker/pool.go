package worker

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Job represents a work item
type Job struct {
	ID       string
	Type     string
	Payload  map[string]interface{}
	Priority int
	Retry    int
	MaxRetry int
	Created  time.Time
	Started  time.Time
	Finished time.Time
	Error    error
	Result   interface{}
}

// JobHandler defines the interface for job processing
type JobHandler interface {
	Handle(ctx context.Context, job *Job) error
	Type() string
}

// Pool represents a worker pool
type Pool struct {
	workers     []*Worker
	jobQueue    chan *Job
	resultQueue chan *Job
	handlers    map[string]JobHandler
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	size        int
	
	// Statistics
	totalJobs     int64
	completedJobs int64
	failedJobs    int64
	activeJobs    int64
	
	// Configuration
	queueSize   int
	maxRetries  int
	jobTimeout  time.Duration
	
	mu sync.RWMutex
}

// Worker represents a single worker
type Worker struct {
	id       int
	pool     *Pool
	jobQueue chan *Job
	quit     chan bool
	active   bool
	mu       sync.RWMutex
}

// NewPool creates a new worker pool
func NewPool(size int) *Pool {
	if size <= 0 {
		size = runtime.NumCPU()
	}

	return &Pool{
		size:        size,
		queueSize:   1000,
		maxRetries:  3,
		jobTimeout:  30 * time.Second,
		jobQueue:    make(chan *Job, 1000),
		resultQueue: make(chan *Job, 1000),
		handlers:    make(map[string]JobHandler),
		workers:     make([]*Worker, size),
	}
}

// RegisterHandler registers a job handler
func (p *Pool) RegisterHandler(handler JobHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[handler.Type()] = handler
}

// Start starts the worker pool
func (p *Pool) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	
	// Start workers
	for i := 0; i < p.size; i++ {
		worker := &Worker{
			id:       i,
			pool:     p,
			jobQueue: p.jobQueue,
			quit:     make(chan bool),
		}
		p.workers[i] = worker
		p.wg.Add(1)
		go worker.start()
	}
	
	// Start result processor
	p.wg.Add(1)
	go p.processResults()
	
	log.Printf("Worker pool started with %d workers", p.size)
}

// Stop stops the worker pool
func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	
	// Stop all workers
	for _, worker := range p.workers {
		if worker != nil {
			worker.stop()
		}
	}
	
	// Wait for all workers to finish
	p.wg.Wait()
	
	// Close channels
	close(p.jobQueue)
	close(p.resultQueue)
	
	log.Printf("Worker pool stopped")
}

// Submit submits a job to the pool
func (p *Pool) Submit(job *Job) error {
	if p.ctx == nil {
		return fmt.Errorf("worker pool not started")
	}
	
	// Set job defaults
	if job.ID == "" {
		job.ID = fmt.Sprintf("job_%d_%d", time.Now().UnixNano(), atomic.AddInt64(&p.totalJobs, 1))
	}
	if job.MaxRetry == 0 {
		job.MaxRetry = p.maxRetries
	}
	job.Created = time.Now()
	
	// Check if handler exists
	p.mu.RLock()
	_, exists := p.handlers[job.Type]
	p.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("no handler registered for job type: %s", job.Type)
	}
	
	// Submit job
	select {
	case p.jobQueue <- job:
		atomic.AddInt64(&p.totalJobs, 1)
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetStats returns pool statistics
func (p *Pool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"workers":        p.size,
		"queue_size":     len(p.jobQueue),
		"queue_capacity": cap(p.jobQueue),
		"total_jobs":     atomic.LoadInt64(&p.totalJobs),
		"completed_jobs": atomic.LoadInt64(&p.completedJobs),
		"failed_jobs":    atomic.LoadInt64(&p.failedJobs),
		"active_jobs":    atomic.LoadInt64(&p.activeJobs),
		"handlers":       len(p.handlers),
	}
}

// Size returns the number of workers
func (p *Pool) Size() int {
	return p.size
}

// processResults processes completed jobs
func (p *Pool) processResults() {
	defer p.wg.Done()
	
	for {
		select {
		case job := <-p.resultQueue:
			if job.Error != nil {
				atomic.AddInt64(&p.failedJobs, 1)
				log.Printf("Job %s failed: %v", job.ID, job.Error)
				
				// Retry if possible
				if job.Retry < job.MaxRetry {
					job.Retry++
					job.Error = nil
					job.Started = time.Time{}
					job.Finished = time.Time{}
					
					select {
					case p.jobQueue <- job:
						log.Printf("Job %s retrying (attempt %d/%d)", job.ID, job.Retry, job.MaxRetry)
					default:
						log.Printf("Failed to requeue job %s for retry", job.ID)
						atomic.AddInt64(&p.failedJobs, 1)
					}
				}
			} else {
				atomic.AddInt64(&p.completedJobs, 1)
				log.Printf("Job %s completed successfully in %v", job.ID, job.Finished.Sub(job.Started))
			}
			
			atomic.AddInt64(&p.activeJobs, -1)
			
		case <-p.ctx.Done():
			return
		}
	}
}

// Worker methods

// start starts the worker
func (w *Worker) start() {
	defer w.pool.wg.Done()
	
	log.Printf("Worker %d started", w.id)
	
	for {
		select {
		case job := <-w.jobQueue:
			w.processJob(job)
			
		case <-w.quit:
			log.Printf("Worker %d stopped", w.id)
			return
			
		case <-w.pool.ctx.Done():
			log.Printf("Worker %d stopped (context cancelled)", w.id)
			return
		}
	}
}

// stop stops the worker
func (w *Worker) stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.active {
		close(w.quit)
		w.active = false
	}
}

// processJob processes a single job
func (w *Worker) processJob(job *Job) {
	w.mu.Lock()
	w.active = true
	w.mu.Unlock()
	
	defer func() {
		w.mu.Lock()
		w.active = false
		w.mu.Unlock()
	}()
	
	atomic.AddInt64(&w.pool.activeJobs, 1)
	job.Started = time.Now()
	
	log.Printf("Worker %d processing job %s (type: %s)", w.id, job.ID, job.Type)
	
	// Get handler
	w.pool.mu.RLock()
	handler, exists := w.pool.handlers[job.Type]
	w.pool.mu.RUnlock()
	
	if !exists {
		job.Error = fmt.Errorf("no handler for job type: %s", job.Type)
		job.Finished = time.Now()
		w.pool.resultQueue <- job
		return
	}
	
	// Create job context with timeout
	ctx, cancel := context.WithTimeout(w.pool.ctx, w.pool.jobTimeout)
	defer cancel()
	
	// Process job
	if err := handler.Handle(ctx, job); err != nil {
		job.Error = err
	}
	
	job.Finished = time.Now()
	
	// Send result
	select {
	case w.pool.resultQueue <- job:
	case <-w.pool.ctx.Done():
		return
	}
}

// IsActive returns whether the worker is currently processing a job
func (w *Worker) IsActive() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.active
}

// GetID returns the worker ID
func (w *Worker) GetID() int {
	return w.id
}
