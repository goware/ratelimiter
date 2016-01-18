package ratelimiter

import (
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

// IsAllowed returns true if the lock allows one more event to happen.
func (lim *Lock) IsAllowed() bool {
	if lim.allowed > 0 {
		hits, err := lim.s.Get(lim.key)
		if err != nil {
			// Something happened with the store, instead of getting crazy about it
			// we simply let the request pass.
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
