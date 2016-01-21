package ratelimiter

import (
	"errors"
	"sync"
	"time"
)

type mockStoreValue struct {
	value     uint64
	expiresAt int64
}

type mockStore struct {
	mu     sync.Mutex
	values map[string]*mockStoreValue
}

func newMockStore() (*mockStore, error) {
	return &mockStore{values: make(map[string]*mockStoreValue)}, nil
}

func (ms *mockStore) Delete(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.values, key)
	return nil
}

func (ms *mockStore) InitWithTTL(key string, ttl int64) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.values[key] = &mockStoreValue{
		expiresAt: time.Now().Unix() + ttl,
	}
	return nil
}

func (ms *mockStore) Get(key string) (uint64, error) {
	value, err := ms.getKey(key)
	if err != nil {
		return 0, err
	}
	return value.value, nil
}

func (ms *mockStore) Increment(key string) (uint64, error) {
	value, err := ms.getKey(key)
	if err != nil {
		return 0, err
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	value.value++
	return value.value, nil
}

func (ms *mockStore) GetTTL(key string) (int64, error) {
	value, err := ms.getKey(key)
	if err != nil {
		return 0, err
	}
	return value.expiresAt - time.Now().Unix(), nil
}

func (ms *mockStore) getKey(key string) (*mockStoreValue, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	v, ok := ms.values[key]
	if !ok {
		return nil, errors.New("No such key")
	}

	ttl := v.expiresAt - time.Now().Unix()
	if ttl < 0 {
		delete(ms.values, key)
		return nil, errors.New("No such key")
	}

	return v, nil
}
