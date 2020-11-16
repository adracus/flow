// Package flow provides concurrency primitives for Go.
//
// The core assumption is that all functions suffice the `Func` type
// definition. Currently, if you ever require to handle moving values
// between the functions, you should make use of proper 'closurization'.
package flow

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Func is a context-aware computation that may produce an error.
type Func func(context.Context) error

// StringFunc is a context-aware computation that may produce an error or a string.
type StringFunc func(context.Context) (string, error)

// IntFunc is a context-aware computation that may produce an error or an int.
type IntFunc func(context.Context) (int, error)

// BoolFunc is a context-aware computation that may produce an error or a bool.
type BoolFunc func(context.Context) (bool, error)

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

func (m multiError) ErrorOrNil() error {
	if len(m) > 0 {
		return m
	}
	return nil
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

type Flow struct {
	executor Executor
}

func New(executor Executor) *Flow {
	return &Flow{executor}
}

func (f *Flow) runAll(l int, run func(i int), deferred func()) {
	if l == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(l)
	for i := 0; i < l; i++ {
		i := i
		f.executor.Submit(func() {
			defer wg.Done()
			run(i)
		})
	}

	go func() {
		defer deferred()
		wg.Wait()
	}()
}

// Parallel runs the given functions in parallel.
//
// It collects all the errors in the returned error. To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) Parallel(ctx context.Context, fns ...Func) error {
	if len(fns) == 0 {
		return nil
	}

	results := make(chan error)
	f.runAll(len(fns), func(i int) {
		results <- fns[i](ctx)
	}, func() { close(results) })

	var errs multiError
	for err := range results {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

// ParallelCancelOnError runs the given functions in parallel, cancelling all if one fails.
//
// It collects all the errors in the returned error. To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelCancelOnError(ctx context.Context, fns ...Func) error {
	if len(fns) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error)
	f.runAll(len(fns), func(i int) {
		err := fns[i](ctx)
		results <- err
	}, func() { close(results) })

	var errs multiError
	for err := range results {
		if err != nil {
			cancel()
			errs = append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

// Race runs all functions in parallel and returns the first that completes.
//
// Completion means a function either errors or succeeds.
// The result of the succeeded function is returned, the other results are
// discarded.
func (f *Flow) Race(ctx context.Context, fns ...Func) error {
	if len(fns) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error)
	f.runAll(len(fns), func(i int) {
		results <- fns[i](ctx)
	}, func() { close(results) })

	err := <-results
	cancel()
	for range results {
	}
	return err
}

type stringResult struct {
	item string
	err  error
}

// ParallelString runs the given functions in parallel.
//
// It collects all the errors and results (regardless if there were errors or not). To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelString(ctx context.Context, fns ...StringFunc) ([]string, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	c := make(chan stringResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		c <- stringResult{item, err}
	}, func() { close(c) })

	var (
		out  []string
		errs multiError
	)
	for res := range c {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		out = append(out, res.item)
	}
	return out, errs.ErrorOrNil()
}

// ParallelStringCancelOnError runs the given functions in parallel, cancelling all if one fails.
//
// It collects all the errors and results (regardless if there were errors or not). To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelStringCancelOnError(ctx context.Context, fns ...StringFunc) ([]string, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan stringResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		c <- stringResult{item, err}
	}, func() { close(c) })

	var (
		out  []string
		errs multiError
	)
	for res := range c {
		if res.err != nil {
			cancel()
			errs = append(errs, res.err)
			continue
		}
		out = append(out, res.item)
	}
	return out, errs.ErrorOrNil()
}

// RaceString runs all functions in parallel and returns the results of the first that completes.
//
// Completion means a function either errors or succeeds.
// The result of the succeeded function is returned, the other results are
// discarded.
func (f *Flow) RaceString(ctx context.Context, fns ...StringFunc) (string, error) {
	if len(fns) == 0 {
		return "", nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan stringResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		results <- stringResult{item, err}
	}, func() { close(results) })

	res := <-results
	cancel()
	for range results {
	}
	return res.item, res.err
}

type intResult struct {
	item int
	err  error
}

// ParallelInt runs the given functions in parallel.
//
// It collects all the errors and results (regardless if there were errors or not). To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelInt(ctx context.Context, fns ...IntFunc) ([]int, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	c := make(chan intResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		c <- intResult{item, err}
	}, func() { close(c) })

	var (
		out  []int
		errs multiError
	)
	for res := range c {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		out = append(out, res.item)
	}
	return out, errs.ErrorOrNil()
}

// ParallelIntCancelOnError runs the given functions in parallel, cancelling all if one fails.
//
// It collects all the errors and results (regardless if there were errors or not). To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelIntCancelOnError(ctx context.Context, fns ...IntFunc) ([]int, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan intResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		c <- intResult{item, err}
	}, func() { close(c) })

	var (
		out  []int
		errs multiError
	)
	for res := range c {
		if res.err != nil {
			cancel()
			errs = append(errs, res.err)
			continue
		}
		out = append(out, res.item)
	}
	return out, errs.ErrorOrNil()
}

// RaceInt runs all functions in parallel and returns the results of the first that completes.
//
// Completion means a function either errors or succeeds.
// The result of the succeeded function is returned, the other results are
// discarded.
func (f *Flow) RaceInt(ctx context.Context, fns ...IntFunc) (int, error) {
	if len(fns) == 0 {
		return 0, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan intResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		results <- intResult{item, err}
	}, func() { close(results) })

	res := <-results
	cancel()
	for range results {
	}
	return res.item, res.err
}

type boolResult struct {
	item bool
	err  error
}

// ParallelInt runs the given functions in parallel.
//
// It collects all the errors and results (regardless if there were errors or not). To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelBool(ctx context.Context, fns ...BoolFunc) ([]bool, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	c := make(chan boolResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		c <- boolResult{item, err}
	}, func() { close(c) })

	var (
		out  []bool
		errs multiError
	)
	for res := range c {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		out = append(out, res.item)
	}
	return out, errs.ErrorOrNil()
}

// ParallelBoolCancelOnError runs the given functions in parallel, cancelling all if one fails.
//
// It collects all the errors and results (regardless if there were errors or not). To obtain
// the multiple errors, use the `Errors` function.
func (f *Flow) ParallelBoolCancelOnError(ctx context.Context, fns ...BoolFunc) ([]bool, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan boolResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		c <- boolResult{item, err}
	}, func() { close(c) })

	var (
		out  []bool
		errs multiError
	)
	for res := range c {
		if res.err != nil {
			cancel()
			errs = append(errs, res.err)
			continue
		}
		out = append(out, res.item)
	}
	return out, errs.ErrorOrNil()
}

// RaceBool runs all functions in parallel and returns the results of the first that completes.
//
// Completion means a function either errors or succeeds.
// The result of the succeeded function is returned, the other results are
// discarded.
func (f *Flow) RaceBool(ctx context.Context, fns ...BoolFunc) (bool, error) {
	if len(fns) == 0 {
		return false, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan boolResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		results <- boolResult{item, err}
	}, func() { close(results) })

	res := <-results
	cancel()
	for range results {
	}
	return res.item, res.err
}

// RaceCond runs all functions in parallel and returns the result of the first function that completes with an
// error or with a truthy result.
func (f *Flow) RaceCond(ctx context.Context, fns ...BoolFunc) (bool, error) {
	if len(fns) == 0 {
		return false, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan boolResult)
	f.runAll(len(fns), func(i int) {
		item, err := fns[i](ctx)
		results <- boolResult{item, err}
	}, func() { close(results) })

	var out boolResult
	for res := range results {
		if res.err != nil || res.item {
			cancel()
			out = res
			break
		}
	}
	for range results {
	}
	return out.item, out.err
}
