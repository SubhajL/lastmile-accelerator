package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"example.com/lma/observability-service/internal/repository"
	"example.com/lma/observability-service/internal/services"
)

type SLOEvaluator struct {
	repo repository.SLORepository
	svc  services.SLOService
}

func NewSLOEvaluator(repo repository.SLORepository, svc services.SLOService) *SLOEvaluator {
	return &SLOEvaluator{repo: repo, svc: svc}
}

// Run starts a periodic evaluation loop with a worker pool. Returns when ctx is canceled.
func (e *SLOEvaluator) Run(ctx context.Context, interval time.Duration, workers int) error {
	if workers < 1 {
		workers = 1
	}
	// Initial pass on start
	_ = e.runOnce(ctx)
	if interval <= 0 {
		return nil
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_ = e.runOnce(ctx)
		}
	}
}

// runOnce evaluates all active SLOs sequentially.
func (e *SLOEvaluator) runOnce(ctx context.Context) error {
	start := time.Now()
	log.Println("scheduler: tick start")
	slos, err := e.repo.ListAll(ctx)
	if err != nil {
		// record duration even on error
		getRecorder().ObserveDuration(time.Since(start))
		getRecorder().SetLastRun(time.Now())
		return err
	}
	ids := make([]string, 0, len(slos))
	for _, s := range slos {
		ids = append(ids, s.ID)
	}
	err = e.runBatch(ctx, ids, 1)
	getRecorder().ObserveDuration(time.Since(start))
	getRecorder().SetLastRun(time.Now())
	log.Printf("scheduler: tick end duration=%s", time.Since(start))
	return err
}

// runBatch evaluates the provided SLO IDs using up to `workers` goroutines.
func (e *SLOEvaluator) runBatch(ctx context.Context, sloIDs []string, workers int) error {
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []string

	worker := func() {
		defer wg.Done()
		for id := range jobs {
			if ctx.Err() != nil {
				return
			}
			if _, err := e.svc.EvaluateSLO(ctx, id); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s: %v", id, err))
				mu.Unlock()
			}
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}
	for _, id := range sloIDs {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return ctx.Err()
		case jobs <- id:
		}
	}
	close(jobs)
	wg.Wait()

	// record metrics
	errCount := len(errs)
	if n := len(sloIDs); n > 0 { getRecorder().IncEvaluated(n) }
	if errCount > 0 { getRecorder().IncErrors(errCount) }
	log.Printf("scheduler: evaluated=%d errors=%d", len(sloIDs), errCount)
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
