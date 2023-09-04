package tests

import "time"

// NewChannelWithTimeout creates a new channel that will be closed after a timeout.
// This function is useful for synchronizing the execution of asynchronous assertion functions, such as Eventually().
func NewChannelWithTimeout(timeout time.Duration) chan struct{} {
	done := make(chan struct{}, 1)
	_ = time.AfterFunc(timeout, func() {
		close(done)
	})

	return done
}
