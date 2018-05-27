package datastore

import "errors"

// DS global datastore
var ds Datastore

// errors
var (
	// ErrNotInitialized (ds == nil)
	ErrNotInitialized = errors.New("datastore: service not initialized")

	// ErrNotFound is returned by Get, Delete if key is not found
	ErrNotFound = errors.New("datastore: key not found")
)

// Datastore represents the storage interface
type Datastore interface {
	put(key string, value []byte) error

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
