package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var data = []byte("hello peerpx")
var key = "peerpxKey"

func TestDatastoreFile(t *testing.T) {
	// not a valid path
	_, err := NewDatastoreFs("/foo")
	assert.Error(t, err)
	// valid path
	ds, err := NewDatastoreFs("/tmp")
	if assert.NoError(t, err) {
		// Put
		err = ds.Put(key, data)
		assert.NoError(t, err)

		// Get
		rData, err := ds.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, string(data), string(rData))

		// Delete
		err = ds.Delete(key)
		assert.NoError(t, err)
		_, err = ds.Get(key)
		assert.Equal(t, ErrNotFoundInDatastore, err)
	}

}
