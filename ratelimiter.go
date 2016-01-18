// package ratelimiter provides functions for counting a number events within a
// given timeframe.
package ratelimiter

import (
	"math"
	"time"
)

// RateLimiter represents a rate limiter.
type RateLimiter struct {
	Store
}

// NewRateLimiter expects a store function and returns a rate limited backed by
// it.
func NewRateLimiter(storeFn StoreFn) (*RateLimiter, error) {
	store, err := storeFn()
	if err != nil {
		return nil, err
	}
	return &RateLimiter{Store: store}, nil
}

// NewLock creates a lock that will allow a given number of events to happen
// before returning an error
func (rl *RateLimiter) NewLock(key string, allowed uint64, duration time.Duration) (*Lock, error) {
	secs := int64(math.Ceil(duration.Seconds()))

	lim := &Lock{
		s:       rl.Store,
		key:     key,
		allowed: allowed,
		ttl:     secs,
	}

	d, err := lim.s.GetTTL(lim.key)

	if err != nil || d < 0 {
		// Either there is no key or the key has expired, so we create a new one
		// with the given lifetime.
		if err := lim.s.InitWithTTL(lim.key, secs); err != nil {
			return nil, err
		}
	}

	return lim, nil
}

// RemoveLock deletes the lock given its key.
func (rl *RateLimiter) RemoveLock(key string) error {
	return rl.Store.Delete(key)
}
