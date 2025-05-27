package utils

import (
	"context"
	"fmt"
	"time"
)

type RetryableFunc[T any] func() (T, error)

// RetryWrapper executes a retryable function with a context and retries on error.
// We could also take a request builder function as an argument and only return the response object
func RetryWrapper[T any](ctx context.Context, fn RetryableFunc[T]) (ret T, err error) {

	ret, err = fn()
	if err == nil {
		return ret, nil // Success
	}

	for _, delaySeconds := range Cfg.Server.RetryFrequencySec {
		select {
		case <-ctx.Done():
			return ret, ctx.Err()
		case <-time.After(time.Duration(delaySeconds) * time.Second):
			// Retry the function
			ret, err = fn()
			if err == nil {
				return ret, nil
			}
		}
	}

	return ret, fmt.Errorf("all retries failed, last error: %w", err)
}
