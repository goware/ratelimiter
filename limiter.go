// package ratelimiter provides functions for counting a number events within a
// given timeframe.
package ratelimiter

import (
	"errors"
	"math"
	"time"
)

// Lock represents a security lock that allows a certain number of
// actions to happen without a fixed time frame.
type Lock struct {
	key     string
	allowed uint64
	ttl     int64
	s       Store
}

// RateLimiter creates a lock that will allow a given number of events to
// happen before locking itself; if a previous lick with the same key already
// exists it will return it instead of creating a new one.
func RateLimiter(key string, allowed uint64, duration time.Duration) (*Lock, error) {
	if DefaultStore == nil {
		return nil, errors.New(`Undefined DefaultStore().`)
	}

	secs := int64(math.Ceil(duration.Seconds()))

	s, err := DefaultStore()
	if err != nil {
		return nil, err
	}

	lim := &Lock{
		s:       s,
		key:     key,
		allowed: allowed,
		ttl:     secs,
	}

	d, err := lim.s.GetTTL(lim.key)

	if err != nil || d < 0 {
		// Either there is no key or the key has expired.
		if err := lim.s.InitWithTTL(lim.key, secs); err != nil {
			return nil, err
		}
	}

	return lim, nil
}

// Remove deletes the lock given its key.
func Remove(key string) error {
	s, err := DefaultStore()
	if err != nil {
		return err
	}
	return s.Delete(key)
}

// IsAllowed returns true if the lock allows one more event to happen.
func (lim *Lock) IsAllowed() bool {
	if lim.allowed > 0 {
		hits, err := lim.s.Get(lim.key)
		if err != nil {
			// Something happened with the store, instead of getting crazy about it
			// we simply let the request everything pass.
			return true
		}
		if hits < lim.allowed {
			// We haven't hit the limit yet.
			return true
		}
		return false
	}
	// No limit has been set.
	return true
}

// Hit logs one event.
func (lim *Lock) Hit() error {
	_, err := lim.s.Increment(lim.key)
	return err
}

// GetTTL returns the number of seconds this key is still valid or zero if the
// key does not exists.
func (lim *Lock) GetTTL() (time.Duration, error) {
	d, err := lim.s.GetTTL(lim.key)
	if err != nil {
		return 0, nil
	}
	return time.Second * time.Duration(d), nil
}

// SetStore sets the store the rate limiter is going to use.
func SetStore(fn StoreFn) {
	if fn == nil {
		panic("Store cannot be nil.")
	}
	DefaultStore = fn
}
