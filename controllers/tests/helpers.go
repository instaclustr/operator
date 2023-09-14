package tests

import (
	"time"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
)

// NewChannelWithTimeout creates a new channel that will be closed after a timeout.
// This function is useful for synchronizing the execution of asynchronous assertion functions, such as Eventually().
func NewChannelWithTimeout(timeout time.Duration) chan struct{} {
	done := make(chan struct{}, 1)
	_ = time.AfterFunc(timeout, func() {
		close(done)
	})

	return done
}

func removeUserByIndex(s []*v1beta1.UserReference, index int) []*v1beta1.UserReference {
	return append(s[:index], s[index+1:]...)
}
