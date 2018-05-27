package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var value = []byte("hello peerpx")
var key = "peerpxKey"

func TestErrNotInitialized(t *testing.T) {
	err := Put(key, value)
	if assert.Error(t, err) {
		assert.EqualError(t, ErrNotInitialized, err.Error())
	}
}

func TestInitFilesystemDatastore(t *testing.T) {
	// not a valid path
	err := InitFilesystemDatastore("/foo")
	assert.Error(t, err)
	// valid path
	err = InitFilesystemDatastore("/tmp")
	if assert.NoError(t, err) {
		// Put
		err = Put(key, value)
		assert.NoError(t, err)

		// Get
		rData, err := Get(key)
		assert.NoError(t, err)
		assert.Equal(t, string(value), string(rData))

		// Delete
		err = Delete(key)
		assert.NoError(t, err)
		_, err = Get(key)
		assert.Equal(t, ErrNotFound, err)
	}
}

func TestInitMokedDatastore(t *testing.T) {
	// it useless to test err in this case
	InitMokedDatastore(value, nil)
	v, err := Get("key")
	if assert.NoError(t, err) {
		assert.Equal(t, value, v)
	}

	// test error
	InitMokedDatastore(value, ErrNotFound)
	_, err = Get(key)
	assert.Error(t, err)
	assert.EqualError(t, ErrNotFound, err.Error())

}
