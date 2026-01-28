// Package goroutine provides utilities for running goroutines with panic recovery.
// This ensures that if any goroutine panics, the error is captured and propagated
// to the error channel, allowing the application to shut down gracefully rather
// than leaving zombie goroutines running.
package goroutine

import (
	"fmt"
	"runtime/debug"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
)

// RunFunc is a function that can be run in a goroutine.
type RunFunc func() error

// Run executes the given function in a new goroutine with panic recovery.
// If the function panics, the panic is recovered, logged with a stack trace,
// and converted to an error that is sent to the error channel.
//
// This is similar to Rust's tokio task spawning with JoinHandle, where panics
// are propagated rather than silently killing the goroutine.
//
// Usage:
//
//	errCh := make(chan error)
//	goroutine.Run(errCh, logger, "http-server", func() error {
//	    return server.ListenAndServe(ctx)
//	})
func Run(errCh chan<- error, logger logging.Logger, name string, fn RunFunc) {
	go func() {
		var err error

		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())

				// Log the panic with full stack trace
				logger.
					WithField("panic", r).
					WithField("stack", stack).
					Error().
					Message(fmt.Sprintf("goroutine '%s' panicked", name))

				// Convert panic to error
				err = errors.Wrap(
					fmt.Errorf("panic: %v\n\nstack trace:\n%s", r, stack),
					fmt.Sprintf("goroutine '%s' panicked", name),
				)
			}

			// Always send to error channel (either the function's error or panic error)
			errCh <- errors.Wrap(err, fmt.Sprintf("%s error", name))
		}()

		err = fn()
	}()
}

// RunNamed is a convenience wrapper that creates a named logger for the goroutine.
func RunNamed(errCh chan<- error, logger logging.Logger, name string, fn RunFunc) {
	Run(errCh, logger.New(name), name, fn)
}
