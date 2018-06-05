package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	assert.NoError(t, InitBasicCache())
	cacheTest(t)
}
