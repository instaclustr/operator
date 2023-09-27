package ratelimiter

import (
	"math"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
)

var (
	DefaultBaseDelay = 1 * time.Minute
	DefaultMaxDelay  = 10 * time.Minute
	defaultMaxTries  = 3
)

type ItemExponentialFailureRateLimiterWithMaxTries struct {
	// failuresLock mutex, which is necessary due to the fact that controller-runtime allows for concurrent reconciler invocations for a single controller
	failuresLock sync.Mutex
	failures     map[interface{}]int
	maxTries     int

	baseDelay time.Duration
	maxDelay  time.Duration
}

func NewItemExponentialFailureRateLimiterWithMaxTries(baseDelay time.Duration, maxDelay time.Duration) ratelimiter.RateLimiter {
	return &ItemExponentialFailureRateLimiterWithMaxTries{
		failures:  map[interface{}]int{},
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		maxTries:  defaultMaxTries,
	}
}

func (r *ItemExponentialFailureRateLimiterWithMaxTries) When(item interface{}) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	backoff := float64(r.baseDelay.Nanoseconds()) * math.Pow(2, float64(r.failures[item]))
	durationBackoff := time.Duration(backoff)

	r.failures[item] = r.failures[item] + 1

	if r.failures[item] > r.maxTries {
		// reminder
		return 60 * time.Minute
	}

	if durationBackoff > r.maxDelay {
		return r.maxDelay
	}

	return durationBackoff
}

func (r *ItemExponentialFailureRateLimiterWithMaxTries) NumRequeues(item interface{}) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

func (r *ItemExponentialFailureRateLimiterWithMaxTries) Forget(item interface{}) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.failures, item)
}
