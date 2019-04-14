package flow

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Func is a context-aware computation that may produce an error.
type Func func(context.Context) error

type multiError []error

// Error implements error.
func (m multiError) Error() string {
	var (
		buf   strings.Builder
		first = true
	)
	for _, err := range m {
		if !first {
			_, _ = fmt.Fprintln(&buf)
		}
		first = false
		buf.WriteString(err.Error())
	}
	return buf.String()
}

// Errors retrieves all causes of a parallel execution.
func Errors(err error) []error {
	if m, ok := err.(multiError); ok {
		return m
	}
	return nil
}

// Sequence runs the given computations one after another.
//
// If one of the functions fails, the sequence stops immediately and the error
// is returned.
// If the context expires between the functions, the context error is returned.
func Sequence(ctx context.Context, fns ...Func) error {
	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

// Parallel runs the given functions in parallel.
//
// It collects all the errors in the returned error. To obtain
// the multiple errors, use the `Errors` function.
func Parallel(ctx context.Context, fns ...Func) error {
	if len(fns) == 0 {
		return nil
	}

	var (
		errors = make(chan error)
		wg     sync.WaitGroup
	)

	wg.Add(len(fns))
	go func() {
		defer close(errors)
		wg.Wait()
	}()

	for _, fn := range fns {
		go func(fn Func) {
			defer wg.Done()
			errors <- fn(ctx)
		}(fn)
	}

	var m multiError
	for err := range errors {
		if err != nil {
			m = append(m, err)
		}
	}
	if len(m) > 0 {
		return m
	}
	return nil
}

// Race runs all functions in parallel and returns the first that completes.
//
// Completion means a function either errors or succeeds.
// The result of the succeeded function is returned, the other results are
// discarded.
func Race(ctx context.Context, fns ...Func) error {
	if len(fns) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		errors = make(chan error)
		wg     sync.WaitGroup
	)

	wg.Add(len(fns))
	go func() {
		defer close(errors)
		wg.Wait()
	}()

	for _, fn := range fns {
		go func(fn Func) {
			defer wg.Done()
			err := fn(ctx)
			errors <- err
			if err != nil {
				cancel()
			}
		}(fn)
	}

	err := <-errors
	go func() {
		for range errors {
		}
	}()
	return err
}
