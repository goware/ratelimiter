# ratelimiter

Package ratelimiter provides functions for counting a number events within a
given timeframe.

```go
import "github.com/goware/ratelimiter"
```

`ratelimiter` depends on an external storage to be able to save its state, this
external storage can be any struct that satisfies the following interface:

```go
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
```

The actual implementation of the store depends on the user. If you want to see
a working example of a redis-based store to start with, check the
`redis_store_test.go` file.

## Passing the store to ratelimiter

`ratelimiter` does not know anything about your package or about your store, in
order to make `ratelimiter` use a store you use the `SetStore()` function.

```
ratelimiter.SetStore(appRateLimitStore)
```

`SetStore()` accepts a function that returns a ready-to-user `Store` or an
error.

## Setting and using locks

Locks are tied to events and events are defined by a name. You can use any name
you want as long as the store supports it.

```go
limiter, err := ratelimiter.RateLimiter("login-attempt-from-127.0.0.1", allowedTries, timeFrame)
```

Whenever the event you defined happens use the `Hit()` method to record the
event and increment the internal event counter:

```go
limiter.Hit()
```

If you want to know if the number of hits is lower than the allowed number of
hits, you can use `IsAllowed()`:

```go
if limiter.IsAllowed() {
	// Yes, one more time, no problem.
}
```

## Example: limiting the number of logins

This is a simple example that shows how to allow 5 login attempts within 10
minutes.

```go
func Login() error {
	// The key can be based on the ID of the user, the IP or the client or
	// whatever makes sense for your app.
	key := getUniqueKeyForThisClient()

	// five attempts within ten minutes.
	lim, _ := ratelimiter.RateLimiter(key, 5, time.Minute*10)

	if lim.IsAllowed() {
		// Recoding login attempt, you can also try to record only failed login
		// attempts.
		lim.Hit()

		// Do some stuff...

		if userCanLogin() {
			// Removing lock after a succesful login so the user does not get blocked
			// if she tries to log in again.
			lim.Remove(key)

			// Do some stuff...
			return  nil
		}

		return errors.New("Sorry. Login failed.")
	}

	return errors.New("Sorry. You have failed too many times!")
}
```
