package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// this will test config && basicconfig

func TestBasic(t *testing.T) {
	// init from empty reader
	// test get && set
	if assert.NoError(t, InitBasicConfig(strings.NewReader(""))) {
		assert.NoError(t, Set("string", "value"))
		assert.Panics(t, func() { Set("intasstring", 5) })
		assert.NoError(t, Set("int", "6"))
		assert.Equal(t, "value", GetString("string"))
		assert.Equal(t, 6, GetInt("int"))
	}

	// init from file
	if assert.NoError(t, InitBasicConfigFromFile("../../etc/samples/config_basic.test")) {
		configTest(t)
	}

}
