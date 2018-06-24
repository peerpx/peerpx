package datastore

import "errors"

// DS global datastore
var ds Provider

// errors
var (
	// ErrNotInitialized (ds == nil)
	ErrNotInitialized = errors.New("datastore: service not initialized")

	// ErrNotFound is returned by Get, Delete if key is not found
	ErrNotFound = errors.New("datastore: key not found")
)

// Provider represents the storage interface interafce
type Provider interface {
	put(key string, value []byte) error

	exists(key string) (bool, error)

	get(key string) (value []byte, err error)

	delete(key string) error
}

// Put store value identified by key
func Put(key string, value []byte) error {
	if ds == nil {
		return ErrNotInitialized
	}
	return ds.put(key, value)
}

func Exists(key string) (bool, error) {
	if ds == nil {
		return false, ErrNotInitialized
	}
	return ds.exists(key)
}

// Get return value associated with key
func Get(key string) (value []byte, err error) {
	if ds == nil {
		return value, ErrNotInitialized
	}
	return ds.get(key)
}

// Delete deletes value for a given key
func Delete(key string) error {
	if ds == nil {
		return ErrNotInitialized
	}
	return ds.delete(key)
}
