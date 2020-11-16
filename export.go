package flow

var (
	// Default returns the default *Flow with an UnlimitedExecutor.
	Default = New(UnlimitedExecutor)

	// Parallel runs the given functions in parallel.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	Parallel = Default.Parallel
	// ParallelCancelOnError runs the given functions in parallel, cancelling all if one fails.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelCancelOnError = Default.ParallelCancelOnError
	// Race runs all functions in parallel and returns the first that completes.
	//
	// Completion means a function either errors or succeeds.
	// The result of the succeeded function is returned, the other results are
	// discarded.
	Race = Default.Race

	// ParallelString runs the given functions in parallel.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelString = Default.ParallelString
	// ParallelStringCancelOnError runs the given functions in parallel, cancelling all if one fails.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelStringCancelOnError = Default.ParallelStringCancelOnError
	// RaceString runs all functions in parallel and returns the first that completes.
	//
	// Completion means a function either errors or succeeds.
	// The result of the succeeded function is returned, the other results are
	// discarded.
	RaceString = Default.RaceString

	// ParallelInt runs the given functions in parallel.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelInt = Default.ParallelInt
	// ParallelIntCancelOnError runs the given functions in parallel, cancelling all if one fails.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelIntCancelOnError = Default.ParallelIntCancelOnError
	// RaceInt runs all functions in parallel and returns the first that completes.
	//
	// Completion means a function either errors or succeeds.
	// The result of the succeeded function is returned, the other results are
	// discarded.
	RaceInt = Default.RaceInt

	// ParallelBool runs the given functions in parallel.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelBool = Default.ParallelBool
	// ParallelBoolCancelOnError runs the given functions in parallel, cancelling all if one fails.
	//
	// It collects all the errors in the returned error. To obtain
	// the multiple errors, use the `Errors` function.
	ParallelBoolCancelOnError = Default.ParallelBoolCancelOnError
	// RaceBool runs all functions in parallel and returns the first that completes.
	//
	// Completion means a function either errors or succeeds.
	// The result of the succeeded function is returned, the other results are
	// discarded.
	RaceBool = Default.RaceBool

	// RaceCond runs all functions in parallel and returns the result of the first function that completes with an
	// error or with a truthy result.
	RaceCond = Default.RaceCond
)
