package ratelimiter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestSetDefaultStore(t *testing.T) {
	// External packages that import the rate limiter are expected to implement a
	// Store and use the SetStore method to let the rate limiter which store to
	// use.
	SetStore(exampleStoreFn)
	assert.NotNil(t, DefaultStore)
}

func TestRemoveKey(t *testing.T) {
	// We want to start clean, so we delete the key we're going to use in the
	// next test.
	err := Remove("login-attempt-from-127.0.0.1")
	assert.Nil(t, err)
}

func TestCheckFiveAttemptsWithinTenSeconds(t *testing.T) {
	allowedTries := uint64(5)

	// Setting up a lock that counts how many login attempts are recevied from
	// 127.0.0.1, we want this rate limiter to lock itself if more than 5 hits
	// happen within 10 seconds.
	rl, err := RateLimiter("login-attempt-from-127.0.0.1", allowedTries, time.Second*10)
	assert.Nil(t, err)

	for i := uint64(0); i < allowedTries*1000; i++ {
		if rl.IsAllowed() {
			assert.False(t, i >= allowedTries, fmt.Errorf("Expecting %d to be lower than %d.", i, allowedTries))
		} else {
			assert.False(t, i < allowedTries, fmt.Errorf("Expecting %d to be greater than or equal to %d.", i, allowedTries))
		}
		err := rl.Hit()
		assert.NoError(t, err)
	}
}

func TestOneMoreTimeWithTheSameKey(t *testing.T) {
	allowedTries := uint64(5)

	rl, err := RateLimiter("login-attempt-from-127.0.0.1", allowedTries, time.Second*10)
	assert.NoError(t, err)

	assert.False(t, rl.IsAllowed())
}

func TestALotMoreTimesWithTheSameKey(t *testing.T) {
	allowedTries := uint64(5)

	for i := 0; i < 1000; i++ {
		rl, err := RateLimiter("login-attempt-from-127.0.0.1", allowedTries, time.Second*10)
		assert.NoError(t, err)
		assert.False(t, rl.IsAllowed(), "This key should still be not allowed.")
	}
}

func TestWithGoroutines(t *testing.T) {
	var wg sync.WaitGroup

	key := "test-reset-password-attempt-from-127.0.0.1"

	err := Remove(key)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			rl, err := RateLimiter(key, 10, time.Minute*1)
			if err != nil {
				panic(err.Error())
			}
			if !rl.IsAllowed() {
				panic("Expecting it to be allowed.")
			}
			rl.Hit()
			wg.Done()
		}(&wg)
	}

	wg.Wait()

	// Another one should trigger the alarm.
	rl, err := RateLimiter(key, 10, time.Minute*1)
	rl.Hit()
	assert.NoError(t, err)
	assert.False(t, rl.IsAllowed(), "Should not be allowed.")
}

func TestWithClear(t *testing.T) {
	key := "test-reset-password-attempt-from-127.0.0.1"

	for i := 0; i < 100; i++ {
		Remove(key)
		rl, err := RateLimiter(key, 10, time.Minute*1)
		assert.NoError(t, err)
		assert.True(t, rl.IsAllowed())

		err = rl.Hit()
		assert.NoError(t, err)
	}
}

func TestFillItAndWaitAFewSeconds(t *testing.T) {
	key := "test-login-attempt-from-127.0.0.1"
	blockDuration := time.Second * 10

	Remove(key)
	for i := 0; i < 100; i++ {
		rl, err := RateLimiter(key, 10, blockDuration)
		assert.NoError(t, err)
		if i >= 10 {
			assert.False(t, rl.IsAllowed(), "The limit have been hit!")
		} else {
			assert.True(t, rl.IsAllowed())
		}
		err = rl.Hit()
		assert.NoError(t, err)
	}

	startTime := time.Now()

	rl, err := RateLimiter(key, 10, blockDuration)
	assert.NoError(t, err)

	// OK, let's keep bugging it until it let's us pass.
	for {
		if rl.IsAllowed() {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}

	endTime := time.Now()

	actualBlockDuration := endTime.Sub(startTime)

	assert.False(t, (blockDuration-actualBlockDuration) > time.Millisecond*100, "The difference between the expected block time and the actual block time should be close to zero.")
}
