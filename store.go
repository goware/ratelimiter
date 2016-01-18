package ratelimiter

// Store interface represents methods to be implemented by a store that can be
// used by a RateLimiter.
type Store interface {
	// Delete removes the key given its name.
	Delete(key string) error

	// InitWithTTL creates a key with the given name, sets it to zero and sets
	// the time to live (in seconds).
	InitWithTTL(key string, ttl int64) error

	// Get returns the numeric value of the key.
	Get(key string) (uint64, error)

	// GetTTL returns the number of seconds before the key expires.
	GetTTL(key string) (int64, error)

	// Increment atomically increments the given key by one.
	Increment(key string) (uint64, error)
}

// StoreFn returns a Store.
type StoreFn func() (Store, error)

// DefaultStore is a function that returns a store defined outside this
// package.
var DefaultStore StoreFn
