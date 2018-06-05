package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func cacheTest(t *testing.T) {
	// set
	assert.NoError(t, Set("foo", []byte("bar")))
	// get
	v, err := Get("foo")
	if assert.NoError(t, err) {
		assert.Equal(t, []byte("bar"), v)
	}
	// del
	assert.NoError(t, Del("foo"))
	// get not found
	v, err = Get("foo")
	if assert.EqualError(t, err, ErrNotFound.Error()) {
		assert.Nil(t, v)
	}
}

func TestNotInitialized(t *testing.T) {
	cache = nil
	_, err := Get("foo")
	assert.EqualError(t, err, ErrNotInitialized.Error())
	assert.EqualError(t, Set("foo", []byte("bar")), ErrNotInitialized.Error())
	assert.EqualError(t, Del("foo"), ErrNotInitialized.Error())
}
