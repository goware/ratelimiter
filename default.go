package ratelimiter

import (
	"errors"
	"time"
)

var errNoSuchDefaultRateLimiter = errors.New("No such rate limiter.")

var defaultRateLimiter *RateLimiter

// Creates the default rate limiter.
func SetStore(storeFn StoreFn) (err error) {
	defaultRateLimiter, err = NewRateLimiter(storeFn)
	return
}

// NewLock creates a lock on the default rate limiter.
func NewLock(key string, allowed uint64, duration time.Duration) (*Lock, error) {
	if defaultRateLimiter == nil {
		return nil, errNoSuchDefaultRateLimiter
	}
	return defaultRateLimiter.NewLock(key, allowed, duration)
}

// RemoveLock executes RemoveLock on the default rate limiter.
func RemoveLock(key string) error {
	if defaultRateLimiter == nil {
		return errNoSuchDefaultRateLimiter
	}
	return defaultRateLimiter.RemoveLock(key)
}
