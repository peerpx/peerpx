package cache

import "errors"

var (
	ErrNotFound       = errors.New("cache: key not found")
	ErrNotInitialized = errors.New("cache: service not initialized")
)

var cache CacheProvider

// CacheProvider
type CacheProvider interface {
	get(key string) ([]byte, error)
	set(key string, value []byte) error
	del(key string) error
}

// Get
func Get(key string) ([]byte, error) {
	if cache == nil {
		return nil, ErrNotInitialized
	}
	return cache.get(key)
}

// Set
func Set(key string, value []byte) error {
	if cache == nil {
		return ErrNotInitialized
	}
	return cache.set(key, value)
}

// Del
func Del(key string) error {
	if cache == nil {
		return ErrNotInitialized
	}
	return cache.del(key)
}
