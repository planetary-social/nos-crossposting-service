package goroutine

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
)

func TestRun_NormalError(t *testing.T) {
	errCh := make(chan error, 1)
	logger := logging.NewDevNullLogger()
	expectedErr := errors.New("test error")

	Run(errCh, logger, "test-runner", func() error {
		return expectedErr
	})

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "test error") {
			t.Fatalf("expected error to contain 'test error', got: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestRun_PanicRecovery(t *testing.T) {
	errCh := make(chan error, 1)
	logger := logging.NewDevNullLogger()

	Run(errCh, logger, "panicking-runner", func() error {
		panic("test panic")
	})

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error from panic, got nil")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "panic") {
			t.Fatalf("expected error to contain 'panic', got: %v", errStr)
		}
		if !strings.Contains(errStr, "test panic") {
			t.Fatalf("expected error to contain 'test panic', got: %v", errStr)
		}
		if !strings.Contains(errStr, "stack trace") {
			t.Fatalf("expected error to contain 'stack trace', got: %v", errStr)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for panic error")
	}
}

func TestRun_NilError(t *testing.T) {
	errCh := make(chan error, 1)
	logger := logging.NewDevNullLogger()

	Run(errCh, logger, "ok-runner", func() error {
		return nil
	})

	select {
	case err := <-errCh:
		// errors.Wrap(nil, ...) returns nil, which is correct behavior
		// The goroutine completed successfully
		if err != nil {
			t.Fatalf("expected nil error for successful completion, got: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}
}
