// nolint:revive
package util

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrWaitTimeout is returned when the WaitGroup fails to complete within the specified duration.
	ErrWaitTimeout = errors.New("WaitGroup has timed out")
)

// WaitTimeout waits for the provided WaitGroup to complete or for the timeout to expire.
//
// Note: If the timeout occurs, the goroutine waiting on wg.Wait() will continue to run in the background (leaked) until it completes.
//
// Parameters:
//   - wg: The WaitGroup to wait for.
//   - timeout: The maximum duration to wait.
//
// Returns:
//   - error: nil if the WaitGroup completed, or ErrWaitTimeout if the timeout was reached.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	waitChan := make(chan struct{})
	go func() {
		defer close(waitChan)
		wg.Wait()
	}()
	select {
	case <-waitChan:
		return nil // completed normally
	case <-time.After(timeout):
		return ErrWaitTimeout // timed out
	}
}
