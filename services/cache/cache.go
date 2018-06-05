package cache

import "errors"

var (
	ErrNotFound       = errors.New("cache: key not found")
	ErrNotInitialized = errors.New("cache: service not initialized")
)

var cache Provider

// Provider interface representing cache provider
type Provider interface {
	get(key string) ([]byte, error)
	set(key string, value []byte) error
	del(key string) error
}

// Get returns value associated with key or error
func Get(key string) ([]byte, error) {
	if cache == nil {
		return nil, ErrNotInitialized
	}
	return cache.get(key)
}

// Set put value associated with key key in cache
func Set(key string, value []byte) error {
	if cache == nil {
		return ErrNotInitialized
	}
	return cache.set(key, value)
}

// Del delete key->value from cache
func Del(key string) error {
	if cache == nil {
		return ErrNotInitialized
	}
	return cache.del(key)
}
