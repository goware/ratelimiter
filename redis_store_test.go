package ratelimiter

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"sync"
)

type redisStore struct {
	c  redis.Conn
	mu sync.Mutex
}

func newRedisStore(addr string) (*redisStore, error) {
	c, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &redisStore{c: c}, nil
}

func (rs *redisStore) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	// This prevent redis from getting crazy with concurrency, see go test -v -race.
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.c.Do(commandName, args...)
}

func (rs *redisStore) Delete(key string) error {
	_, err := rs.do("DEL", key)
	return err
}

func (rs *redisStore) InitWithTTL(key string, ttl int64) (err error) {
	if _, err = rs.do("SET", key, 0, "EX", ttl); err != nil {
		return err
	}
	return nil
}

func (rs *redisStore) Get(key string) (uint64, error) {
	b, err := rs.do("GET", key)
	if err != nil {
		return 0, err
	}
	if b != nil {
		i, err := strconv.Atoi(string(b.([]byte)))
		return uint64(i), err
	}
	return 0, nil
}

func (rs *redisStore) GetTTL(key string) (int64, error) {
	i, err := rs.do("TTL", key)
	if err != nil {
		return 0, err
	}
	return i.(int64), err
}

func (rs *redisStore) Increment(key string) (uint64, error) {
	i, err := rs.do("INCR", key)
	if err != nil {
		return 0, err
	}
	return uint64(i.(int64)), err
}
