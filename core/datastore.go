package core

import "errors"

// DS global datastore
var DS Datastore

// Datastore represents the storage interface
type Datastore interface {
	// Put store value identified by key
	Put(key string, value []byte) error

	// Get return value associated with key
	Get(key string) (value []byte, err error)

	// Delete deletes value for a given key
	Delete(key string) error
}

// ErrNotFoundInDatastore is returned by Get, Delete if key is not found
var ErrNotFoundInDatastore = errors.New("datastore: key not found")
