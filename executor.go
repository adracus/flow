package flow

import (
	"fmt"
	"sync"
)

// Executor allows non-blocking submission of functions.
type Executor interface {
	// Submit schedules f for execution in a non-blocking way.
	Submit(f func())
}

type plainExecutor struct{}

func (plainExecutor) Submit(f func()) {
	go f()
}

// UnlimitedExecutor is an Executor that dispatches every function immediately with `go func()`.
var UnlimitedExecutor Executor = plainExecutor{}

// LimitingExecutor represents a pool of goroutines.
type LimitingExecutor struct {
	maxRunning int
	executor   Executor
	lock       sync.Mutex

	running bool
	ingest  chan<- func()
}

// LimitExecutor creates a new Executor with the given maximum number of goroutines that may run simultaneously.
func LimitExecutor(limit int, executor Executor) *LimitingExecutor {
	if limit < 0 {
		panic(fmt.Errorf("limit may not be < 0 but was %d", limit))
	}
	return &LimitingExecutor{maxRunning: limit, executor: executor}
}

// Start launches the pool, making it ready to accept submissions.
func (p *LimitingExecutor) Start() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.ingest == nil {
		var (
			ingest = make(chan func())
			queue  []func()
		)
		p.ingest = ingest
		go func() {
			var (
				current int
				wg      sync.WaitGroup
				done    = make(chan struct{})
			)
			defer close(done)

		Loop:
			for {
				select {
				case <-done:
					current--
				case f, ok := <-ingest:
					if !ok {
						break Loop
					}
					queue = append(queue, f)
				default:
					if len(queue) > 0 && current < p.maxRunning {
						current++
						f := queue[0]
						queue = queue[1:]
						wg.Add(1)
						p.executor.Submit(func() {
							defer wg.Done()
							f()
							done <- struct{}{}
						})
					}
				}
			}

			wg.Wait()
		}()
	}
}

// Submit schedules f to be executed in a non-blocking way.
func (p *LimitingExecutor) Submit(f func()) {
	p.ingest <- f
}

// Stop stops the executor. Goroutines that already were running will continue to run, unless cancelled otherwise.
func (p *LimitingExecutor) Stop() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.ingest != nil {
		close(p.ingest)
		p.ingest = nil
	}
}
