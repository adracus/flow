package flow

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// Race yields the result of the 'winner', all other functions are canceled.
func ExampleRace() {
	var (
		f1 = func(ctx context.Context) error {
			fmt.Println("Won the race!")
			return nil
		}
		f2 = func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return nil
			}
		}
	)

	if err := Race(context.Background(), f1, f2); err != nil {
		log.Fatalf("Error occurred: %v", err)
	}
	// Output: Won the race!
}

// Parallel runs the functions in parallel.
func ExampleParallel() {
	var (
		f1 = func(ctx context.Context) error {
			fmt.Println("f1 running")
			return nil
		}
		f2 = func(ctx context.Context) error {
			fmt.Println("f2 running")
			return nil
		}
		f3 = func(ctx context.Context) error { return errors.New("error") }
	)

	if err := Parallel(context.Background(), f1, f2, f3); err != nil {
		fmt.Println(err.Error())
	}
	// Unordered Output: f1 running
	// f2 running
	// error
}
